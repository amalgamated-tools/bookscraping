package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
)

//go:embed all:dist
var distFS embed.FS

// Server represents the HTTP server with embedded frontend
type Server struct {
	addr     string
	queries  *db.Queries
	grClient *goodreads.Client
	blClient *booklore.Client
	mux      *http.ServeMux
	eventCh  chan string
}

// NewServer creates a new server instance
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		eventCh: make(chan string, 100),
	}
	for _, opt := range opts {
		opt(s)
	}

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

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/config", s.handleGetConfig)
	s.mux.HandleFunc("POST /api/config", s.handleSaveConfig)
	s.mux.HandleFunc("POST /api/testConnection", s.handleTestConnection)

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

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start starts the server on the given address
func (s *Server) Start() error {
	slog.Info("Starting server", "address", s.addr)
	return http.ListenAndServe(s.addr, s)
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
