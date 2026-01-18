package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

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
	bookIDToDBID := make(map[int64]int64) // Map book.ID to insertedBook.ID

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

		// Store the mapping for later use in series linking
		bookIDToDBID[book.ID] = insertedBook.ID

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

	// Sync unique series and link books to series, and series to authors
	seriesNameToID := make(map[string]int64)
	for seriesName := range uniqueSeries {
		series, err := s.queries.UpsertSeries(ctx, db.UpsertSeriesParams{
			SeriesID:    0, // SeriesID is not available from Booklore, we get it from goodreads
			Name:        seriesName,
			Description: nil,
			Url:         nil,
			Data:        nil,
		})
		if err != nil {
			slog.Warn("Failed to upsert series during sync", "series_name", seriesName, "error", err)
			continue
		}
		seriesNameToID[seriesName] = series.ID
	}

	// Second pass: link books to series and extract series authors
	for _, book := range books {
		if book.SeriesName == "" {
			continue
		}

		seriesID, exists := seriesNameToID[book.SeriesName]
		if !exists {
			continue
		}

		// Get book database ID from the mapping created in first pass
		dbBookID, exists := bookIDToDBID[book.ID]
		if !exists {
			slog.Error("Failed to find book ID mapping", "book_id", book.ID)
			continue
		}

		err = s.queries.UpdateBookSeries(ctx, db.UpdateBookSeriesParams{
			SeriesID: &seriesID,
			ID:       dbBookID,
		})
		if err != nil {
			slog.Error("Failed to link book to series", "book_id", book.ID, "series_id", seriesID, "error", err)
			continue
		}

		// Link authors to series
		for _, authorName := range book.Authors {
			author, err := s.queries.GetAuthorByName(ctx, authorName)
			if err != nil {
				// Author should already exist from the previous pass, but just in case
				author, err = s.queries.UpsertAuthor(ctx, authorName)
				if err != nil {
					slog.Error("Failed to upsert author", "author_name", authorName, "error", err)
					continue
				}
			}

			err = s.queries.LinkSeriesAuthor(ctx, db.LinkSeriesAuthorParams{
				SeriesID: seriesID,
				AuthorID: author.ID,
			})
			if err != nil {
				slog.Error("Failed to link series author", "series_id", seriesID, "author", authorName, "error", err)
			}
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
