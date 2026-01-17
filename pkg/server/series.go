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

// handleGetSeriesFromGoodreads fetches series data from Goodreads and creates missing books
func (s *Server) handleGetSeriesFromGoodreads(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	seriesID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	ctx := context.Background()

	// Get the series info
	if _, err := s.queries.GetSeries(ctx, seriesID); err != nil {
		slog.Error("Failed to get series", "id", seriesID, "error", err)
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	// Get at least one book from this series to extract the Goodreads series ID
	books, err := s.queries.GetBooksBySeries(ctx, &seriesID)
	if err != nil || len(books) == 0 {
		slog.Error("No books found for series", "series_id", seriesID)
		writeError(w, http.StatusBadRequest, "Series has no books - cannot fetch from Goodreads")
		return
	}

	// Find a book with a Goodreads ID to extract the series from
	var goodreadsSeriesID string
	for _, book := range books {
		if book.GoodreadsID != nil && *book.GoodreadsID != "" {
			// TODO: Fetch book from Goodreads and extract series ID
			// For now, we'll need the series ID to be in the books data
			slog.Info("Found book with Goodreads ID", "goodreads_id", *book.GoodreadsID)
			// This would require scraping the Goodreads book page to find the series
			break
		}
	}

	// If we don't have a series ID yet, return an error
	if goodreadsSeriesID == "" {
		slog.Warn("Could not find Goodreads series ID from books")
		writeError(w, http.StatusBadRequest, "Unable to find Goodreads series ID - none of the books have Goodreads IDs")
		return
	}

	slog.Info("Would fetch Goodreads series", "series_id", goodreadsSeriesID)
	writeJSON(w, map[string]string{"status": "success", "message": "Goodreads series fetch would be implemented"})
}
