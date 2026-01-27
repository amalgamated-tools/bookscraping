package server

import (
	"context"
	"encoding/json"
	"fmt"
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
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			slog.Error("Failed to decode credentials from request body", slog.Any("error", err))
		}
	}

	s.emitEvent("sync_started", "Starting synchronization...", map[string]any{})

	// Use provided creds or fall back to stored config or env
	var client *booklore.Client

	// Try to get config from DB first
	storedServerUrl, _ := s.queries.GetConfig(ctx, "serverUrl")
	storedUsername, _ := s.queries.GetConfig(ctx, "username")
	storedPassword, _ := s.queries.GetConfig(ctx, "password")

	// Precedence: Request Body > DB Config > Env Vars (via initial client)
	if creds.ServerURL != "" && creds.Username != "" && creds.Password != "" {
		client = booklore.NewClient(
			booklore.WithBaseURL(creds.ServerURL),
			booklore.WithCredentials(creds.Username, creds.Password),
		)
	} else if storedServerUrl != "" && storedUsername != "" && storedPassword != "" {
		client = booklore.NewClient(
			booklore.WithBaseURL(storedServerUrl),
			booklore.WithCredentials(storedUsername, storedPassword),
		)
	} else if os.Getenv("BOOKLORE_SERVER") != "" {
		client = s.blClient
	} else {
		s.emitEvent("sync_error", "Booklore credentials required", map[string]any{})
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
			s.emitEvent("sync_progress", "Authenticated with stored token", map[string]any{})
		} else {
			slog.Info("Stored token invalid, attempting fresh login")
			s.emitEvent("sync_progress", "Stored token invalid, logging in again...", map[string]any{})
			// Token is invalid, fall through to login
			if err := client.Login(ctx); err != nil {
				slog.Error("Failed to login to Booklore", slog.Any("error", err))
				s.emitEvent("sync_error", "Failed to login to Booklore", map[string]any{})
				writeError(w, http.StatusUnauthorized, "Failed to login to Booklore")
				return
			}
			// Store the new token
			newToken := client.GetToken()
			err = s.queries.SetConfig(ctx, db.SetConfigParams{
				Key:   "booklore_access_token",
				Value: newToken.AccessToken,
			})
			if err != nil {
				slog.Error("Failed to store new access token", slog.Any("error", err))
			}
		}
	} else {
		// No token stored, perform login
		s.emitEvent("sync_progress", "No token found, logging in to Booklore...", map[string]any{})
		if err := client.Login(ctx); err != nil {
			slog.Error("Failed to login to Booklore", slog.Any("error", err))
			s.emitEvent("sync_error", "Failed to login to Booklore", map[string]any{})
			writeError(w, http.StatusUnauthorized, "Failed to login to Booklore")
			return
		}
		// Store the token
		token := client.GetToken()
		err := s.queries.SetConfig(ctx, db.SetConfigParams{
			Key:   "booklore_access_token",
			Value: token.AccessToken,
		})
		if err != nil {
			slog.Error("Failed to store new access token", slog.Any("error", err))
		}
		if token.RefreshToken != "" {
			err := s.queries.SetConfig(ctx, db.SetConfigParams{
				Key:   "booklore_refresh_token",
				Value: token.RefreshToken,
			})
			if err != nil {
				slog.Error("Failed to store new refresh token", slog.Any("error", err))
			}
		}
	}

	// Fetch books
	s.emitEvent("sync_progress", "Fetching books from Booklore...", map[string]any{})
	books, err := client.LoadAllBooks()
	if err != nil {
		slog.Error("Failed to fetch books from Booklore", slog.Any("error", err))
		s.emitEvent("sync_error", "Failed to fetch books from Booklore", map[string]any{})
		writeError(w, http.StatusInternalServerError, "Failed to fetch books")
		return
	}

	slog.Info("Fetched books from Booklore", slog.Int("count", len(books)))
	s.emitEvent("sync_progress", fmt.Sprintf("Fetched %d books from Booklore", len(books)), map[string]any{"books_count": len(books)})

	// Sync books to DB
	s.emitEvent("sync_progress", "Starting to sync books to database...", map[string]any{})
	syncedCount := 0
	uniqueSeries := make(map[string]struct{})
	bookIDToDBID := make(map[int64]int64) // Map book.ID to insertedBook.ID

	for i, book := range books {
		asin := &book.ASIN
		isbn10 := &book.ISBN10
		isbn13 := &book.ISBN13
		hardcoverID := &book.HardCoverID
		hardcoverBookID := &book.HardCoverBookID
		goodreadsID := &book.GoodreadsId
		googleID := &book.GoogleId

		var seriesNamePtr *string
		if book.SeriesName != "" {
			slog.Info("Syncing book in series", slog.String("book_title", book.Title), slog.String("series_name", book.SeriesName))
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
		jsonData, err := json.Marshal(book)
		if err != nil {
			slog.Error("Failed to marshal book JSON", slog.Int64("book_id", book.ID), slog.String("title", book.Title), slog.Any("error", err))
			continue
		}

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
			slog.Error("Failed to sync book", slog.Int64("book_id", book.ID), slog.String("title", book.Title), slog.Any("error", err))
			continue
		}

		// Store the mapping for later use in series linking
		bookIDToDBID[book.ID] = insertedBook.ID

		// Sync authors
		for _, authorName := range book.Authors {
			author, err := s.queries.UpsertAuthor(ctx, authorName)
			if err != nil {
				slog.Error("Failed to upsert author", slog.String("name", authorName), slog.Any("error", err))
				continue
			}

			err = s.queries.LinkBookAuthor(ctx, db.LinkBookAuthorParams{
				BookID:   insertedBook.ID,
				AuthorID: author.ID,
			})
			if err != nil {
				slog.Error("Failed to link book author", slog.String("book_title", book.Title), slog.String("author", authorName), slog.Any("error", err))
			}
		}

		syncedCount++

		// Emit progress every 10 books
		if (i+1)%10 == 0 || i == len(books)-1 {
			progress := float64((i + 1)) / float64(len(books)) * 100
			s.emitEvent("sync_progress", fmt.Sprintf("Synced %d of %d books", i+1, len(books)), map[string]any{
				"synced_books": i + 1,
				"total_books":  len(books),
				"progress":     progress,
			})
		}
	}

	// Sync unique series and link books to series, and series to authors
	s.emitEvent("sync_progress", "Creating series entries...", map[string]any{})
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
			slog.Warn("Failed to upsert series during sync", slog.String("series_name", seriesName), slog.Any("error", err))
			continue
		}
		seriesNameToID[seriesName] = series.ID
	}
	s.emitEvent("sync_progress", fmt.Sprintf("Created %d series", len(uniqueSeries)), map[string]any{"series_count": len(uniqueSeries)})

	// Second pass: link books to series and extract series authors
	s.emitEvent("sync_progress", "Linking books to series and authors...", map[string]any{})
	for i, book := range books {
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
			slog.Error("Failed to find book ID mapping", slog.Int64("book_id", book.ID))
			continue
		}

		err = s.queries.UpdateBookSeries(ctx, db.UpdateBookSeriesParams{
			SeriesID: &seriesID,
			ID:       dbBookID,
		})
		if err != nil {
			slog.Error("Failed to link book to series", slog.Int64("book_id", book.ID), slog.Int64("series_id", seriesID), slog.Any("error", err))
			continue
		}

		// Link authors to series
		for _, authorName := range book.Authors {
			author, err := s.queries.GetAuthorByName(ctx, authorName)
			if err != nil {
				// Author should already exist from the previous pass, but just in case
				author, err = s.queries.UpsertAuthor(ctx, authorName)
				if err != nil {
					slog.Error("Failed to upsert author", slog.String("author_name", authorName), slog.Any("error", err))
					continue
				}
			}

			err = s.queries.LinkSeriesAuthor(ctx, db.LinkSeriesAuthorParams{
				SeriesID: seriesID,
				AuthorID: author.ID,
			})
			if err != nil {
				slog.Error("Failed to link series author", slog.Int64("series_id", seriesID), slog.String("author", authorName), slog.Any("error", err))
			}
		}

		// Emit progress every 10 books
		if (i+1)%10 == 0 || i == len(books)-1 {
			progress := float64((i + 1)) / float64(len(books)) * 100
			s.emitEvent("sync_progress", fmt.Sprintf("Linked %d of %d books to series", i+1, len(books)), map[string]any{
				"linked_books": i + 1,
				"total_books":  len(books),
				"progress":     progress,
			})
		}
	}

	slog.Info("Sync complete", slog.Int("total_books", len(books)), slog.Int("synced_books", syncedCount), slog.Int("synced_series", len(uniqueSeries)))

	// Emit completion event
	s.emitEvent("sync_complete", "Synchronization completed successfully", map[string]any{
		"total_books":   len(books),
		"synced_books":  syncedCount,
		"synced_series": len(uniqueSeries),
	})

	writeJSON(w, map[string]any{
		"status":        "success",
		"total":         len(books),
		"synced":        syncedCount,
		"synced_series": len(uniqueSeries),
	})
}

// Helper function to emit SSE events
func (s *Server) emitEvent(eventType, message string, data map[string]any) {
	event := map[string]any{
		"type":    eventType,
		"message": message,
	}
	// Merge additional data
	for k, v := range data {
		event[k] = v
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal SSE event", slog.Any("error", err))
		return
	}
	select {
	case s.eventCh <- string(eventJSON):
	default:
		slog.Warn("Failed to send SSE event (channel full)")
	}
}
