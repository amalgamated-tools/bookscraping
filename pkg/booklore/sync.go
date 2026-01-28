package booklore

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

// Sync performs synchronization of book data with the server.
func (c *Client) Sync(ctx context.Context, eventCh chan<- string) error {
	methodLogger := slog.With(slog.String("method", "Sync"), slog.String("package", "booklore"))
	// we need to make sure that we have an access token that hasn't expired
	err := c.ValidateToken(ctx)

	// err can be a generic error or one of ErrNoAccessToken/ErrInvalidToken
	switch err {
	case nil:
		methodLogger.InfoContext(ctx, "Access token valid, proceeding with sync")
	case ErrInvalidToken:
		methodLogger.InfoContext(ctx, "Access token invalid, attempting to refresh token")
		// if we have a refresh token, try to refresh the access token
		if c.accessToken.RefreshToken != "" {
			err = c.RefreshToken(ctx)
			if err != nil {
				methodLogger.ErrorContext(ctx, "Failed to refresh access token during sync", slog.Any("error", err))
				// we can now try to login again
				err = c.performLogin(ctx)
				if err != nil {
					methodLogger.ErrorContext(ctx, "Login failed during sync after refresh token failure", slog.Any("error", err))
					// return a wrapped error
					return fmt.Errorf("failed to login during sync after refresh token failure: %w", err)
				}
				methodLogger.InfoContext(ctx, "Login successful after refresh token failure, access token obtained")
			} else {
				methodLogger.InfoContext(ctx, "Access token refreshed successfully, proceeding with sync")
			}
		} else {
			methodLogger.InfoContext(ctx, "No refresh token available, attempting login to obtain access token")
			err = c.performLogin(ctx)
			if err != nil {
				methodLogger.ErrorContext(ctx, "Login failed during sync", slog.Any("error", err))
				// return a wrapped error
				return fmt.Errorf("failed to login during sync: %w", err)
			}
			methodLogger.InfoContext(ctx, "Login successful, access token obtained")
		}
	case ErrNoAccessToken:
		methodLogger.InfoContext(ctx, "No access token found, attempting login to obtain access token")
		err = c.performLogin(ctx)

		if err != nil {
			methodLogger.ErrorContext(ctx, "Login failed during sync", slog.Any("error", err))
			// return a wrapped error
			return fmt.Errorf("failed to login during sync: %w", err)
		}
		methodLogger.InfoContext(ctx, "Login successful, access token obtained")
	}

	emitEvent(eventCh, "sync_progress", "Fetching books from Booklore...", map[string]any{})

	books, err := c.LoadAllBooks()
	if err != nil {
		methodLogger.ErrorContext(ctx, "Failed to load books during sync", slog.Any("error", err))
		emitEvent(eventCh, "sync_error", "Failed to fetch books from Booklore", map[string]any{})
		return fmt.Errorf("failed to load books during sync: %w", err)
	}

	methodLogger.InfoContext(ctx, "Loaded books from server", slog.Int("book_count", len(books)))
	emitEvent(eventCh, "sync_progress", fmt.Sprintf("Fetched %d books from Booklore", len(books)), map[string]any{"books_count": len(books)})
	syncedCount := 0
	uniqueSeries := make(map[string]struct{})
	bookIDToDBID := make(map[int64]int64) // Map book.ID to insertedBook.ID

	for i, book := range books {
		methodLogger.InfoContext(ctx, "Syncing book", slog.Int("current", i+1), slog.Int("total", len(books)), slog.Int64("book_id", book.ID), slog.String("title", book.Title))
		asin := &book.ASIN
		isbn10 := &book.ISBN10
		isbn13 := &book.ISBN13
		hardcoverID := &book.HardCoverID
		hardcoverBookID := &book.HardCoverBookID
		goodreadsID := &book.GoodreadsId
		googleID := &book.GoogleId

		var seriesNamePtr *string
		if book.SeriesName != "" {
			methodLogger.InfoContext(ctx, "Syncing book in series", slog.String("book_title", book.Title), slog.String("series_name", book.SeriesName))
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
			methodLogger.ErrorContext(ctx, "Failed to marshal book JSON", slog.Int64("book_id", book.ID), slog.String("title", book.Title), slog.Any("error", err))
			continue
		}

		insertedBook, err := c.queries.UpsertBook(ctx, db.UpsertBookParams{
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
			methodLogger.ErrorContext(ctx, "Failed to sync book", slog.Int64("book_id", book.ID), slog.String("title", book.Title), slog.Any("error", err))
			continue
		}

		// Store the mapping for later use in series linking
		bookIDToDBID[book.ID] = insertedBook.ID

		// Sync authors
		for _, authorName := range book.Authors {
			author, err := c.queries.UpsertAuthor(ctx, authorName)
			if err != nil {
				methodLogger.ErrorContext(ctx, "Failed to upsert author", slog.String("name", authorName), slog.Any("error", err))
				continue
			}

			err = c.queries.LinkBookAuthor(ctx, db.LinkBookAuthorParams{
				BookID:   insertedBook.ID,
				AuthorID: author.ID,
			})
			if err != nil {
				methodLogger.ErrorContext(ctx, "Failed to link book author", slog.String("book_title", book.Title), slog.String("author", authorName), slog.Any("error", err))
			}
		}

		syncedCount++
		// Emit progress every 10 books
		if (i+1)%10 == 0 || i == len(books)-1 {
			progress := float64((i + 1)) / float64(len(books)) * 100
			emitEvent(eventCh, "sync_progress", fmt.Sprintf("Synced %d of %d books", i+1, len(books)), map[string]any{
				"synced_books": i + 1,
				"total_books":  len(books),
				"progress":     progress,
			})
		}
		methodLogger.InfoContext(ctx, "Book synced successfully", slog.Int64("book_id", insertedBook.BookID), slog.String("title", insertedBook.Title))
		emitEvent(eventCh, "sync_progress", "Creating series entries...", map[string]any{})
	}

	emitEvent(eventCh, "sync_progress", "Creating series entries...", map[string]any{})
	seriesNameToID := make(map[string]int64)
	for seriesName := range uniqueSeries {
		series, err := c.queries.UpsertSeries(ctx, db.UpsertSeriesParams{
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
	emitEvent(eventCh, "sync_progress", fmt.Sprintf("Created %d series", len(uniqueSeries)), map[string]any{"series_count": len(uniqueSeries)})

	return nil
}

func emitEvent(eventCh chan<- string, eventType, message string, data map[string]any) {
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
	case eventCh <- string(eventJSON):
	default:
		slog.Warn("Failed to send SSE event (channel full)")
	}
}
