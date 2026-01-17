package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

// SeriesWithAuthors wraps a Series with its authors
type SeriesWithAuthors struct {
	*db.Series
	Authors []string `json:"authors"`
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

	// Fetch authors for each series
	seriesWithAuthors := make([]SeriesWithAuthors, len(series))
	for i, singleSeries := range series {
		authors, err := s.queries.GetSeriesAuthors(ctx, singleSeries.ID)
		if err != nil {
			slog.Error("Failed to get authors for series", "series_id", singleSeries.ID, "error", err)
			authors = []db.Author{}
		}

		authorNames := make([]string, len(authors))
		for j, author := range authors {
			authorNames[j] = author.Name
		}

		seriesWithAuthors[i] = SeriesWithAuthors{
			Series:  &series[i],
			Authors: authorNames,
		}
	}

	total, err := s.queries.CountSeries(ctx)
	if err != nil {
		slog.Error("Failed to count series", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to count series")
		return
	}

	writeJSON(w, PaginatedResponse{
		Data:    seriesWithAuthors,
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

	// Fetch authors for the series
	authors, err := s.queries.GetSeriesAuthors(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get authors for series", "series_id", id, "error", err)
		authors = []db.Author{}
	}

	authorNames := make([]string, len(authors))
	for i, author := range authors {
		authorNames[i] = author.Name
	}

	seriesWithAuthors := SeriesWithAuthors{
		Series:  &series,
		Authors: authorNames,
	}

	writeJSON(w, seriesWithAuthors)
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

	books, err := s.queries.GetBooksBySeries(ctx, &id)
	if err != nil {
		slog.Error("Failed to get books for series", "series_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to get books for series")
		return
	}

	// Fetch authors for each book
	booksWithAuthors := make([]BookWithAuthors, len(books))
	for i, book := range books {
		authors, err := s.queries.GetAuthorsForBook(ctx, book.ID)
		if err != nil {
			slog.Error("Failed to get authors for book", "book_id", book.ID, "error", err)
			authors = []db.Author{}
		}

		authorNames := make([]string, len(authors))
		for j, author := range authors {
			authorNames[j] = author.Name
		}

		booksWithAuthors[i] = BookWithAuthors{
			Book:    &books[i],
			Authors: authorNames,
		}
	}

	writeJSON(w, booksWithAuthors)
}

// handleGetSeriesFromGoodreads is a stub for fetching series data from Goodreads
func (s *Server) handleGetSeriesFromGoodreads(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	// TODO: Implement Goodreads series scraping
	// This should:
	// 1. Get at least one book from this series
	// 2. Fetch the book's detail page from Goodreads
	// 3. Extract the series link from the book detail
	// 4. Scrape the series page for all books
	// 5. Update the series record with the Goodreads series ID and description
	// 6. Add any missing books to the database

	slog.Info("Goodreads series fetch stub called", "series_id", id)
	writeJSON(w, map[string]string{"status": "stub", "message": "Goodreads integration not yet implemented"})
}
