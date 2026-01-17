package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

// BookWithAuthors wraps a Book with its authors
type BookWithAuthors struct {
	*db.Book
	Authors []string `json:"authors"`
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

	total, err := s.queries.CountBooks(ctx)
	if err != nil {
		slog.Error("Failed to count books", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to count books")
		return
	}

	writeJSON(w, PaginatedResponse{
		Data:    booksWithAuthors,
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

	// Fetch authors for the book
	authors, err := s.queries.GetAuthorsForBook(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get authors for book", "book_id", id, "error", err)
		authors = []db.Author{}
	}

	authorNames := make([]string, len(authors))
	for i, author := range authors {
		authorNames[i] = author.Name
	}

	bookWithAuthors := BookWithAuthors{
		Book:    &book,
		Authors: authorNames,
	}

	writeJSON(w, bookWithAuthors)
}
