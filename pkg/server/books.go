package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

// handleListBooks returns a paginated list of books with authors
func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	offset := (page - 1) * perPage
	ctx := context.Background()

	// Get total count
	count, err := s.queries.CountBooks(ctx)
	if err != nil {
		slog.Error("Failed to count books", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "Failed to load books")
		return
	}

	// Get paginated books
	books, err := s.queries.ListBooks(ctx, db.ListBooksParams{
		Limit:  int64(perPage),
		Offset: int64(offset),
	})
	if err != nil {
		slog.Error("Failed to list books", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "Failed to load books")
		return
	}

	// Enrich books with authors
	bookResponses := make([]bookResponse, 0, len(books))
	for _, book := range books {
		authors, err := s.queries.GetAuthorsForBook(ctx, book.ID)
		if err != nil {
			slog.Warn("Failed to load authors for book", "book_id", book.ID, slog.Any("error", err))
		}

		authorNames := make([]string, 0, len(authors))
		for _, author := range authors {
			authorNames = append(authorNames, author.Name)
		}

		isMissing := false
		if book.IsMissing != nil {
			isMissing = *book.IsMissing
		}

		bookResponses = append(bookResponses, bookResponse{
			ID:              book.ID,
			BookID:          book.BookID,
			Title:           book.Title,
			Description:     book.Description,
			SeriesName:      book.SeriesName,
			SeriesNumber:    book.SeriesNumber,
			SeriesID:        book.SeriesID,
			ASIN:            book.Asin,
			ISBN10:          book.Isbn10,
			ISBN13:          book.Isbn13,
			Language:        book.Language,
			HardcoverID:     book.HardcoverID,
			HardcoverBookID: book.HardcoverBookID,
			GoodreadsID:     book.GoodreadsID,
			GoogleID:        book.GoogleID,
			Authors:         authorNames,
			IsMissing:       isMissing,
		})
	}

	writeJSON(w, PaginatedResponse{
		Data:    bookResponses,
		Total:   count,
		Page:    page,
		PerPage: perPage,
	})
}

// handleGetBook returns a single book with full details
func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	ctx := context.Background()

	book, err := s.queries.GetBook(ctx, id)
	if err != nil {
		slog.Error("Failed to get book", "id", id, slog.Any("error", err))
		writeError(w, http.StatusNotFound, "Book not found")
		return
	}

	// Get authors
	authors, err := s.queries.GetAuthorsForBook(ctx, book.ID)
	if err != nil {
		slog.Warn("Failed to load authors for book", "book_id", book.ID, slog.Any("error", err))
	}

	authorNames := make([]string, 0, len(authors))
	for _, author := range authors {
		authorNames = append(authorNames, author.Name)
	}

	isMissing := false
	if book.IsMissing != nil {
		isMissing = *book.IsMissing
	}

	response := bookResponse{
		ID:              book.ID,
		BookID:          book.BookID,
		Title:           book.Title,
		Description:     book.Description,
		SeriesName:      book.SeriesName,
		SeriesNumber:    book.SeriesNumber,
		SeriesID:        book.SeriesID,
		ASIN:            book.Asin,
		ISBN10:          book.Isbn10,
		ISBN13:          book.Isbn13,
		Language:        book.Language,
		HardcoverID:     book.HardcoverID,
		HardcoverBookID: book.HardcoverBookID,
		GoodreadsID:     book.GoodreadsID,
		GoogleID:        book.GoogleID,
		Authors:         authorNames,
		IsMissing:       isMissing,
	}

	writeJSON(w, response)
}

// handleGetSeries returns a single series with authors
func (s *Server) handleGetSeries(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	ctx := context.Background()

	series, err := s.queries.GetSeries(ctx, id)
	if err != nil {
		slog.Error("Failed to get series", "id", id, slog.Any("error", err))
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	// Get authors for the series
	authors, err := s.queries.GetSeriesAuthors(ctx, series.ID)
	if err != nil {
		slog.Warn("Failed to get authors for series", "series_id", series.ID, slog.Any("error", err))
		authors = []db.Author{}
	}

	authorNames := make([]string, len(authors))
	for i, author := range authors {
		authorNames[i] = author.Name
	}

	response := SeriesWithAuthors{
		Series:  &series,
		Authors: authorNames,
	}

	writeJSON(w, response)
}

// handleGetSeriesBooks returns all books in a series
func (s *Server) handleGetSeriesBooks(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	ctx := context.Background()

	// Verify series exists
	_, err = s.queries.GetSeries(ctx, id)
	if err != nil {
		slog.Error("Failed to get series", "id", id, slog.Any("error", err))
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	// Get books in series
	bookRows, err := s.queries.GetBooksBySeries(ctx, &id)
	if err != nil {
		slog.Error("Failed to get books for series", "series_id", id, slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "Failed to load books")
		return
	}

	// Enrich books with authors
	bookResponses := make([]bookResponse, 0, len(bookRows))
	for _, book := range bookRows {
		authors, err := s.queries.GetAuthorsForBook(ctx, book.ID)
		if err != nil {
			slog.Warn("Failed to load authors for book", "book_id", book.ID, slog.Any("error", err))
		}

		authorNames := make([]string, 0, len(authors))
		for _, author := range authors {
			authorNames = append(authorNames, author.Name)
		}

		isMissing := false
		if book.IsMissing != nil {
			isMissing = *book.IsMissing
		}

		bookResponses = append(bookResponses, bookResponse{
			ID:              book.ID,
			BookID:          book.BookID,
			Title:           book.Title,
			Description:     book.Description,
			SeriesName:      book.SeriesName,
			SeriesNumber:    book.SeriesNumber,
			SeriesID:        book.SeriesID,
			ASIN:            book.Asin,
			ISBN10:          book.Isbn10,
			ISBN13:          book.Isbn13,
			Language:        book.Language,
			HardcoverID:     book.HardcoverID,
			HardcoverBookID: book.HardcoverBookID,
			GoodreadsID:     book.GoodreadsID,
			GoogleID:        book.GoogleID,
			Authors:         authorNames,
			IsMissing:       isMissing,
		})
	}

	writeJSON(w, bookResponses)
}

type bookResponse struct {
	ID              int64    `json:"id"`
	BookID          int64    `json:"book_id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	SeriesName      *string  `json:"series_name,omitempty"`
	SeriesNumber    *float64 `json:"series_number,omitempty"`
	SeriesID        *int64   `json:"series_id,omitempty"`
	ASIN            *string  `json:"asin,omitempty"`
	ISBN10          *string  `json:"isbn10,omitempty"`
	ISBN13          *string  `json:"isbn13,omitempty"`
	Language        *string  `json:"language,omitempty"`
	HardcoverID     *string  `json:"hardcover_id,omitempty"`
	HardcoverBookID *int64   `json:"hardcover_book_id,omitempty"`
	GoodreadsID     *string  `json:"goodreads_id,omitempty"`
	GoogleID        *string  `json:"google_id,omitempty"`
	Authors         []string `json:"authors,omitempty"`
	IsMissing       bool     `json:"is_missing"`
}
