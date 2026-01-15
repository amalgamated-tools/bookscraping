package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
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
	queries  *db.Queries
	grClient *goodreads.Client
	blClient *booklore.Client
	mux      *http.ServeMux
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
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/books", s.handleListBooks)
	s.mux.HandleFunc("GET /api/books/{id}", s.handleGetBook)
	s.mux.HandleFunc("GET /api/books/search", s.handleSearchBooks)
	s.mux.HandleFunc("GET /api/series", s.handleListSeries)
	s.mux.HandleFunc("POST /api/series/refresh", s.handleRefreshSeries)
	s.mux.HandleFunc("GET /api/series/{id}", s.handleGetSeries)
	s.mux.HandleFunc("POST /api/sync", s.handleSync)

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

// Book handlers
func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	page, perPage := getPagination(r)
	offset := (page - 1) * perPage

	ctx := context.Background()

	books, err := s.queries.ListBooks(ctx, db.ListBooksParams{
		Limit:  int64(perPage),
		Offset: int64(offset),
	})
	if err != nil {
		slog.Error("Failed to list books", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to list books")
		return
	}

	total, err := s.queries.CountBooks(ctx)
	if err != nil {
		slog.Error("Failed to count books", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to count books")
		return
	}

	writeJSON(w, PaginatedResponse{
		Data:    books,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}

func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := s.queries.GetBook(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get book", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "Book not found")
		return
	}

	writeJSON(w, book)
}

func (s *Server) handleSearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "Search query required")
		return
	}

	// For now, just list books - you can implement full-text search later
	books, err := s.queries.ListBooks(context.Background(), db.ListBooksParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		slog.Error("Failed to search books", "error", err)
		writeError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	// Simple title filtering (replace with proper search)
	var filtered []db.Book
	queryLower := strings.ToLower(query)
	for _, book := range books {
		if strings.Contains(strings.ToLower(book.Title), queryLower) {
			filtered = append(filtered, book)
		}
	}

	writeJSON(w, filtered)
}

// Series handlers
func (s *Server) handleListSeries(w http.ResponseWriter, r *http.Request) {
	page, perPage := getPagination(r)
	offset := (page - 1) * perPage

	ctx := context.Background()

	series, err := s.queries.ListSeries(ctx, db.ListSeriesParams{
		Limit:  int64(perPage),
		Offset: int64(offset),
	})
	if err != nil {
		slog.Error("Failed to list series", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to list series")
		return
	}

	total, err := s.queries.CountSeries(ctx)
	if err != nil {
		slog.Error("Failed to count series", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to count series")
		return
	}

	writeJSON(w, PaginatedResponse{
		Data:    series,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}

func (s *Server) handleGetSeries(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	series, err := s.queries.GetSeries(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get series", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	writeJSON(w, series)
}

func (s *Server) handleRefreshSeries(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	series, err := s.queries.GetSeries(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get series", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	go s.refreshSeriesFromBooklore(id, series.Name)

	writeJSON(w, map[string]string{"status": "refreshing"})
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

	// Use provided creds or fall back to env/existing client
	var client *booklore.Client
	if creds.ServerURL != "" && creds.Username != "" && creds.Password != "" {
		client = booklore.NewClient(creds.ServerURL, creds.Username, creds.Password)
	} else if os.Getenv("BOOKLORE_SERVER") != "" {
		client = s.blClient
	} else {
		writeError(w, http.StatusBadRequest, "Booklore credentials required")
		return
	}

	// Ensure logged in
	if err := client.Login(); err != nil {
		slog.Error("Failed to login to Booklore", "error", err)
		writeError(w, http.StatusUnauthorized, "Failed to login to Booklore")
		return
	}

	// Fetch books
	books, err := client.LoadAllBooks()
	if err != nil {
		slog.Error("Failed to fetch books from Booklore", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch books")
		return
	}

	// Sync books to DB
	syncedCount := 0
	uniqueSeries := make(map[int64]string)

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
			seriesNamePtr = &book.SeriesName
		}

		// Collect unique series
		if book.SeriesID != 0 && book.SeriesName != "" {
			uniqueSeries[book.SeriesID] = book.SeriesName
		}

		var seriesNumberPtr *float64
		if book.SeriesNumber != 0 {
			seriesNumberPtr = &book.SeriesNumber
		}

		// Store raw JSON data
		jsonData, _ := json.Marshal(book)

		_, err := s.queries.UpsertBook(ctx, db.UpsertBookParams{
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
		syncedCount++
	}

	// Sync unique series
	for seriesID, seriesName := range uniqueSeries {
		_, err := s.queries.UpsertSeries(ctx, db.UpsertSeriesParams{
			SeriesID:    seriesID,
			Name:        seriesName,
			Description: nil,
			Url:         nil,
			Data:        nil,
		})
		if err != nil {
			slog.Warn("Failed to upsert series during sync", "series_id", seriesID, "error", err)
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

func (s *Server) refreshSeriesFromBooklore(seriesID int64, seriesName string) {
	ctx := context.Background()

	if os.Getenv("BOOKLORE_SERVER") == "" || os.Getenv("BOOKLORE_USERNAME") == "" || os.Getenv("BOOKLORE_PASSWORD") == "" {
		slog.Error("Booklore credentials not configured yet")
		return
	}

	if err := s.blClient.Login(); err != nil {
		slog.Error("Failed to login to Booklore", "error", err)
		return
	}

	books, err := s.blClient.LoadAllBooks()
	if err != nil {
		slog.Error("Failed to load books from Booklore", "error", err)
		return
	}

	matchingBooks := []booklore.Book{}
	for _, book := range books {
		if book.SeriesName == seriesName {
			matchingBooks = append(matchingBooks, book)
		}
	}

	if len(matchingBooks) == 0 {
		slog.Info("No books found for series", "series_name", seriesName)
		return
	}

	for _, book := range matchingBooks {
		asin := &book.ASIN
		isbn10 := &book.ISBN10
		isbn13 := &book.ISBN13
		hardcoverID := &book.HardCoverID
		hardcoverBookID := &book.HardCoverBookID
		goodreadsID := &book.GoodreadsId
		googleID := &book.GoogleId
		seriesNamePtr := &seriesName
		var seriesNumberPtr *float64
		if book.SeriesNumber != 0 {
			seriesNumberPtr = &book.SeriesNumber
		}

		// Also store raw JSON for consistency
		jsonData, _ := json.Marshal(book)

		_, err := s.queries.UpsertBook(ctx, db.UpsertBookParams{
			BookID:          book.ID,
			Title:           book.Title,
			Description:     book.Description,
			SeriesName:      seriesNamePtr,
			SeriesNumber:    seriesNumberPtr,
			Asin:            asin,
			Isbn10:          isbn10,
			Isbn13:          isbn13,
			Language:        nil,
			HardcoverID:     hardcoverID,
			HardcoverBookID: hardcoverBookID,
			GoodreadsID:     goodreadsID,
			GoogleID:        googleID,
			Data:            jsonData,
		})
		if err != nil {
			slog.Error("Failed to upsert book", "book_id", book.ID, "error", err)
			continue
		}
	}

	slog.Info("Successfully refreshed series from Booklore", "series_id", seriesID, "series_name", seriesName, "books_count", len(matchingBooks))
}
