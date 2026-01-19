package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
)

// SeriesWithAuthors wraps a Series with its authors
type SeriesWithAuthors struct {
	*db.Series
	Authors []string `json:"authors"`
}

// SyncSeriesResponse contains the result of syncing a series with Goodreads
type SyncSeriesResponse struct {
	Status          string `json:"status"`
	Message         string `json:"message"`
	SeriesID        int64  `json:"series_id"`
	ExistingBooks   int    `json:"existing_books"`
	MissingBooks    int    `json:"missing_books"`
	NewMissingBooks int    `json:"new_missing_books"`
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
	series, err := s.queries.GetSeries(ctx, seriesID)
	if err != nil {
		slog.Error("Failed to get series", "id", seriesID, "error", err)
		writeError(w, http.StatusNotFound, "Series not found")
		return
	}

	// Get existing books in this series
	existingBooks, err := s.queries.GetBooksBySeries(ctx, &seriesID)
	if err != nil {
		slog.Error("Failed to get books for series", "series_id", seriesID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch series books")
		return
	}

	// Build a set of existing Goodreads IDs to avoid duplicates
	existingGoodreadsIDs := make(map[string]bool)
	for _, book := range existingBooks {
		if book.GoodreadsID != nil && *book.GoodreadsID != "" {
			existingGoodreadsIDs[*book.GoodreadsID] = true
		}
	}

	// Goodreads series ID is stored in the series_id field (as int64, but represents Goodreads ID)
	goodreadsSeriesID := strconv.FormatInt(series.SeriesID, 10)

	// Fetch series from Goodreads
	grClient := goodreads.NewClient()
	booksWithPosition, err := grClient.GetSeriesBooks(goodreadsSeriesID)
	if err != nil {
		slog.Error("Failed to fetch Goodreads series", "goodreads_id", goodreadsSeriesID, "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch from Goodreads")
		return
	}

	slog.Info("Fetched books from Goodreads", "count", len(booksWithPosition), "series_id", goodreadsSeriesID)

	// Create missing books
	newMissingCount := 0
	for _, bp := range booksWithPosition {
		// Skip books we already have
		if existingGoodreadsIDs[bp.Book.BookID] {
			continue
		}

		// Parse series number from position string (e.g., "1", "1.5", "2")
		seriesNumber := 0.0
		if bp.SeriesPosition != "" {
			// Try to parse as float
			if n, err := strconv.ParseFloat(bp.SeriesPosition, 64); err == nil {
				seriesNumber = n
			}
		}

		// Extract plain text description from HTML
		description := bp.Book.Description.TruncatedHTML
		if description == "" {
			description = bp.Book.Description.HTML
		}
		// Basic HTML stripping
		description = stripHTML(description)

		// Create a synthetic book_id from Goodreads ID to ensure uniqueness
		// Use a large offset (10 billion) to avoid conflicts with Booklore IDs
		goodreadsIDNum, _ := strconv.ParseInt(bp.Book.BookID, 10, 64)
		syntheticBookID := 10000000000 + goodreadsIDNum

		// Create the missing book entry
		_, err := s.queries.CreateMissingBook(ctx, db.CreateMissingBookParams{
			BookID:       syntheticBookID,
			Title:        bp.Book.Title,
			Description:  description,
			SeriesName:   &series.Name,
			SeriesNumber: &seriesNumber,
			GoodreadsID:  &bp.Book.BookID,
			SeriesID:     &seriesID,
		})

		if err != nil {
			slog.Error("Failed to create missing book", "book_title", bp.Book.Title, "error", err)
			continue
		}

		slog.Info("Created missing book", "title", bp.Book.Title, "goodreads_id", bp.Book.BookID)
		newMissingCount++
	}

	response := SyncSeriesResponse{
		Status:          "success",
		Message:         "Successfully synced with Goodreads",
		SeriesID:        seriesID,
		ExistingBooks:   len(existingBooks),
		MissingBooks:    len(booksWithPosition),
		NewMissingBooks: newMissingCount,
	}

	writeJSON(w, response)
}

// stripHTML removes HTML tags from a string
func stripHTML(html string) string {
	// Simple HTML tag removal - not production grade
	result := html
	// Remove common HTML tags
	tags := []string{"<p>", "</p>", "<br>", "<br/>", "<div>", "</div>", "<span>", "</span>"}
	for _, tag := range tags {
		result = strings.ReplaceAll(result, tag, " ")
	}
	// Remove any remaining tags
	inTag := false
	var cleaned strings.Builder
	for _, c := range result {
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
		} else if !inTag {
			cleaned.WriteRune(c)
		}
	}
	return strings.TrimSpace(cleaned.String())
}
