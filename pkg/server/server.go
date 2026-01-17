package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
)

//go:embed all:dist
var distFS embed.FS

// Server represents the HTTP server with embedded frontend
type Server struct {
	queries  *db.Queries
	grClient *goodreads.Client
	blClient *booklore.Client
	mux      *http.ServeMux
	eventCh  chan string
}

// NewServer creates a new server instance
func NewServer(queries *db.Queries) *Server {
	blClient := booklore.NewClient(
		os.Getenv("BOOKLORE_SERVER"),
		os.Getenv("BOOKLORE_USERNAME"),
		os.Getenv("BOOKLORE_PASSWORD"),
	)
	s := &Server{
		queries:  queries,
		grClient: goodreads.NewClient(),
		blClient: blClient,
		mux:      http.NewServeMux(),
		eventCh:  make(chan string, 100),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/books", s.handleListBooks)
	s.mux.HandleFunc("GET /api/books/{id}", s.handleGetBook)
	s.mux.HandleFunc("GET /api/series", s.handleListSeries)
	s.mux.HandleFunc("GET /api/series/{id}", s.handleGetSeries)
	s.mux.HandleFunc("POST /api/sync", s.handleSync)
	s.mux.HandleFunc("GET /api/config", s.handleGetConfig)
	s.mux.HandleFunc("POST /api/config", s.handleSaveConfig)
	s.mux.HandleFunc("POST /api/testConnection", s.handleTestConnection)

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
func (s *Server) Start(addr string) error {
	slog.Info("Starting server", "address", addr)
	return http.ListenAndServe(addr, s)
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

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for events
	clientEvents := make(chan string, 10)
	done := make(chan struct{})

	// Set up a ticker to send periodic updates (like a heartbeat)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Handle client disconnect
	go func() {
		<-r.Context().Done()
		close(done)
	}()

	// Send initial connection message
	fmt.Fprintf(w, "data: connected\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	slog.Info("SSE client connected")

	// Subscribe to server events
	go func() {
		for event := range s.eventCh {
			select {
			case <-done:
				return
			case clientEvents <- event:
			}
		}
	}()

	for {
		select {
		case <-done:
			slog.Info("SSE client disconnected")
			return
		case event := <-clientEvents:
			fmt.Fprintf(w, "data: %s\n\n", event)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			slog.Info("Sent SSE event to client", "event", event)
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

func (s *Server) handleTriggerEvent(w http.ResponseWriter, r *http.Request) {
	// Parse event message from request body
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Message == "" {
		writeError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Send the event to all connected SSE clients
	select {
	case s.eventCh <- payload.Message:
		slog.Info("Event triggered", "message", payload.Message)
		writeJSON(w, map[string]string{
			"status":  "success",
			"message": payload.Message,
		})
	case <-time.After(1 * time.Second):
		writeError(w, http.StatusInternalServerError, "Failed to send event")
	}
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse optional credentials from body
	var creds struct {
		ServerURL string `json:"server_url"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&creds)
	}

	// Use provided creds or fall back to stored config or env
	var client *booklore.Client

	// Try to get config from DB first
	storedServerUrl, _ := s.queries.GetConfig(ctx, "serverUrl")
	storedUsername, _ := s.queries.GetConfig(ctx, "username")
	storedPassword, _ := s.queries.GetConfig(ctx, "password")

	// Precedence: Request Body > DB Config > Env Vars (via initial client)
	if creds.ServerURL != "" && creds.Username != "" && creds.Password != "" {
		client = booklore.NewClient(creds.ServerURL, creds.Username, creds.Password)
	} else if storedServerUrl != "" && storedUsername != "" && storedPassword != "" {
		client = booklore.NewClient(storedServerUrl, storedUsername, storedPassword)
	} else if os.Getenv("BOOKLORE_SERVER") != "" {
		client = s.blClient
	} else {
		writeError(w, http.StatusBadRequest, "Booklore credentials required")
		return
	}

	// Try to use stored token if available
	storedAccessToken, _ := s.queries.GetConfig(ctx, "booklore_access_token")
	storedRefreshToken, _ := s.queries.GetConfig(ctx, "booklore_refresh_token")
	if storedAccessToken != "" {
		client.SetToken(booklore.Token{
			AccessToken:  storedAccessToken,
			RefreshToken: storedRefreshToken,
		})
		// Try to validate the token
		if err := client.ValidateToken(); err == nil {
			slog.Info("Using valid stored token")
		} else {
			slog.Info("Stored token invalid, attempting fresh login")
			// Token is invalid, fall through to login
			if err := client.Login(); err != nil {
				slog.Error("Failed to login to Booklore", "error", err)
				writeError(w, http.StatusUnauthorized, "Failed to login to Booklore")
				return
			}
			// Store the new token
			newToken := client.GetToken()
			s.queries.SetConfig(ctx, db.SetConfigParams{
				Key:   "booklore_access_token",
				Value: newToken.AccessToken,
			})
		}
	} else {
		// No token stored, perform login
		if err := client.Login(); err != nil {
			slog.Error("Failed to login to Booklore", "error", err)
			writeError(w, http.StatusUnauthorized, "Failed to login to Booklore")
			return
		}
		// Store the token
		token := client.GetToken()
		s.queries.SetConfig(ctx, db.SetConfigParams{
			Key:   "booklore_access_token",
			Value: token.AccessToken,
		})
		if token.RefreshToken != "" {
			s.queries.SetConfig(ctx, db.SetConfigParams{
				Key:   "booklore_refresh_token",
				Value: token.RefreshToken,
			})
		}
	}

	// Fetch books
	books, err := client.LoadAllBooks()
	if err != nil {
		slog.Error("Failed to fetch books from Booklore", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch books")
		return
	}

	slog.Info("Fetched books from Booklore", "count", len(books))

	// Sync books to DB
	syncedCount := 0
	uniqueSeries := make(map[string]struct{})

	for _, book := range books {
		asin := &book.ASIN
		isbn10 := &book.ISBN10
		isbn13 := &book.ISBN13
		hardcoverID := &book.HardCoverID
		hardcoverBookID := &book.HardCoverBookID
		goodreadsID := &book.GoodreadsId
		googleID := &book.GoogleId

		var seriesNamePtr *string
		if book.SeriesName != "" {
			slog.Info("Syncing book in series", "book_title", book.Title, "series_name", book.SeriesName)
			seriesNamePtr = &book.SeriesName
		}

		// Collect unique series
		if book.SeriesName != "" {
			uniqueSeries[book.SeriesName] = struct{}{}
		}

		var seriesNumberPtr *float64
		if book.SeriesNumber != 0 {
			seriesNumberPtr = &book.SeriesNumber
		}

		// Store raw JSON data
		jsonData, _ := json.Marshal(book)

		insertedBook, err := s.queries.UpsertBook(ctx, db.UpsertBookParams{
			BookID:          book.ID,
			Title:           book.Title,
			Description:     book.Description,
			SeriesName:      seriesNamePtr,
			SeriesNumber:    seriesNumberPtr,
			Asin:            asin,
			Isbn10:          isbn10,
			Isbn13:          isbn13,
			Language:        nil, // Not currently in Book struct
			HardcoverID:     hardcoverID,
			HardcoverBookID: hardcoverBookID,
			GoodreadsID:     goodreadsID,
			GoogleID:        googleID,
			Data:            jsonData,
		})

		if err != nil {
			slog.Error("Failed to sync book", "book_id", book.ID, "title", book.Title, "error", err)
			continue
		}

		// Sync authors
		for _, authorName := range book.Authors {
			author, err := s.queries.UpsertAuthor(ctx, authorName)
			if err != nil {
				slog.Error("Failed to upsert author", "name", authorName, "error", err)
				continue
			}

			err = s.queries.LinkBookAuthor(ctx, db.LinkBookAuthorParams{
				BookID:   insertedBook.ID,
				AuthorID: author.ID,
			})
			if err != nil {
				slog.Error("Failed to link book author", "book_title", book.Title, "author", authorName, "error", err)
			}
		}

		syncedCount++
	}

	// Sync unique series
	for seriesName := range uniqueSeries {
		_, err := s.queries.UpsertSeries(ctx, db.UpsertSeriesParams{
			SeriesID:    0, // SeriesID is not available from Booklore, we get it from goodreads
			Name:        seriesName,
			Description: nil,
			Url:         nil,
			Data:        nil,
		})
		if err != nil {
			slog.Warn("Failed to upsert series during sync", "error", err)
		}
	}

	slog.Info("Sync complete", "total_books", len(books), "synced_books", syncedCount, "synced_series", len(uniqueSeries))
	writeJSON(w, map[string]any{
		"status":        "success",
		"total":         len(books),
		"synced":        syncedCount,
		"synced_series": len(uniqueSeries),
	})
}
