package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

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
