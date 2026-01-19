package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
	"golang.org/x/sync/errgroup"
)

//go:embed all:dist
var distFS embed.FS

const (
	// UserAgentHeader is the header name for the user agent.
	UserAgentHeader = "User-Agent"
	// HTTPWriteTimeout is the maximum duration before timing out writes of the response.
	HTTPWriteTimeout = 10 * time.Second
	// HTTPReadTimeout is the maximum duration for reading the entire request, including the body.
	HTTPReadTimeout = 10 * time.Second
	// HTTPIdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	HTTPIdleTimeout = 30 * time.Second
	// HTTPRequestTimeout is the maximum duration for handling a single HTTP request.
	HTTPRequestTimeout = 10 * time.Second
	// ShutdownGracePeriod is the time we allow for graceful shutdown of the http server
	// Should be longer than HTTPWriteTimeout, but shorter than the k8s terminationGracePeriodSeconds (30 seconds)
	ShutdownGracePeriod = 15 * time.Second
)

// ShutdownFunc is a function that takes a context and returns an error
type ShutdownFunc func(context.Context) error

// Server represents the HTTP server with embedded frontend
type Server struct {
	addr     string
	queries  db.Querier
	grClient *goodreads.Client
	blClient *booklore.Client

	Address string
	port    int

	mux           *http.ServeMux
	httpServer    *http.Server
	shutdownFuncs []ShutdownFunc

	eventCh chan string

	// SSE client tracking
	sseClients map[string]chan string
	sseMu      sync.RWMutex
}

// NewServer creates a new server instance
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		eventCh:    make(chan string, 100),
		sseClients: make(map[string]chan string),
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.port == 0 {
		s.port = 8080
	}
	s.Address = net.JoinHostPort("0.0.0.0", strconv.Itoa(s.port))

	s.setupRoutes()
	s.setupBookloreClient()
	return s
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("Running server", "address", s.Address)
	ctx, cancel := context.WithCancel(ctx)

	timeoutHandler := http.TimeoutHandler(s.mux, HTTPRequestTimeout, "Request timeout")

	s.httpServer = &http.Server{
		Addr:         s.Address,
		Handler:      timeoutHandler,
		WriteTimeout: HTTPWriteTimeout,
		ReadTimeout:  HTTPReadTimeout,
		IdleTimeout:  HTTPIdleTimeout,
	}

	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.Any("error", err))
			s.shutdownFuncs = append(s.shutdownFuncs, func(_ context.Context) error {
				return err
			})
			cancel()
			return
		}
	}()

	s.shutdownFuncs = append(s.shutdownFuncs, s.httpServer.Shutdown)

	<-ctx.Done()
	return s.shutdown(ctx)
}

func (s *Server) shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ShutdownGracePeriod)
	defer cancel()

	shutdownGroup, ctx := errgroup.WithContext(ctx)

	for _, shutdownFn := range s.shutdownFuncs {
		fn := shutdownFn
		shutdownGroup.Go(func() error {
			return fn(ctx)
		})
	}

	return shutdownGroup.Wait()
}

func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/config", s.handleGetConfig)
	s.mux.HandleFunc("POST /api/config", s.handleSaveConfig)

	// Books routes
	s.mux.HandleFunc("GET /api/books", s.handleListBooks)
	s.mux.HandleFunc("GET /api/books/{id}", s.handleGetBook)

	// Series routes
	s.mux.HandleFunc("GET /api/series", s.handleListSeries)
	s.mux.HandleFunc("GET /api/series/{id}", s.handleGetSeries)
	s.mux.HandleFunc("GET /api/series/{id}/books", s.handleGetSeriesBooks)
	s.mux.HandleFunc("POST /api/series/{id}/goodreads", s.handleGetSeriesFromGoodreads)

	s.mux.HandleFunc("POST /api/sync", s.handleSync)

	s.mux.HandleFunc("GET /api/events", s.handleEvents)
	s.mux.HandleFunc("POST /api/events/trigger", s.handleTriggerEvent)

	// Serve embedded frontend
	distContent, err := fs.Sub(distFS, "dist")
	if err != nil {
		slog.Error("Failed to get embedded dist folder", "error", err)
		return
	}

	fileServer := http.FileServer(http.FS(distContent))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		if _, err := fs.Stat(distContent, strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// For SPA routing, serve index.html for non-asset routes
		if !strings.Contains(path, ".") {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		// File not found
		http.NotFound(w, r)
	})
}

func (s *Server) setupBookloreClient() {
	if s.blClient == nil {
		serverURL := ""
		username := ""
		password := ""

		ctx := context.Background()

		// Try to get config from database if queries are available
		if s.queries == nil {
			slog.Warn("No database queries available, BookLore client will have no configuration")
		} else {
			slog.Info("Loading BookLore configuration from database")
			dbServerURL, err := s.queries.GetConfig(ctx, "serverUrl")
			if err == nil && dbServerURL != "" {
				serverURL = dbServerURL
			}
			dbUsername, err := s.queries.GetConfig(ctx, "username")
			if err == nil && dbUsername != "" {
				username = dbUsername
			}
			dbPassword, err := s.queries.GetConfig(ctx, "password")
			if err == nil && dbPassword != "" {
				password = dbPassword
			}
		}

		bookloreClient := booklore.NewClient(serverURL, username, password)
		s.blClient = bookloreClient
	}
}

// JSON response helpers
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Pagination helper
type PaginatedResponse struct {
	Data    any   `json:"data"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

func getPagination(r *http.Request) (page, perPage int) {
	page = 1
	perPage = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}
	return
}
