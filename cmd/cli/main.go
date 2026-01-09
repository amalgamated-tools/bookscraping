package main

import (
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/config"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	grClient := goodreads.NewClient()

	client := booklore.NewClient(cfg.BookloreServer, cfg.BookloreUsername, cfg.BooklorePassword)
	client.Login()

	books, err := client.LoadAllBooks()
	if err != nil {
		panic(err)
	}

	series := make(map[string]string)

	for _, book := range books {
		// does this book have series info?
		if book.SeriesName != "" {
			slog.Info(
				"Book with series",
				slog.String("title", book.Title),
				slog.String("goodreads_id", book.GoodreadsId),
				slog.String("series_name", book.SeriesName),
				slog.Float64("series_number", book.SeriesNumber),
				slog.String("url", fmt.Sprintf("https://booklore.veverka.net/book/%d", book.ID)),
			)
			series[book.SeriesName] = book.GoodreadsId
		}
	}
	slog.Info("Series found", "count", len(series))
	for name, grID := range series {
		slog.Info(" Series", "name", name, "goodreads_id", grID)
		grb, err := grClient.GetBook(grID)
		if err != nil {
			slog.Error(" - error fetching from goodreads", "error", err)
		} else {
			slog.Info(" - found series info", slog.Any("series", grb))
			break
		}
	}
}

// if book.SeriesName != "" {
// 	// slog.Info("Book loaded", "title", book.Title, "series", book.SeriesName, "number", book.SeriesNumber)
// 	if _, ok := series[book.SeriesName]; !ok {
// 		series[book.SeriesName] = make(map[float64]booklore.Book)
// 	}
// 	series[book.SeriesName][book.SeriesNumber] = book
// } else {
// 	slog.Info("Book without series", "title", book.Title)
// 	if book.GoodreadsId != "" {
// 		// slog.Info(" - has goodreads id", "id", book.GoodreadsId)
// 		// grb, err := grClient.GetBook(book.GoodreadsId)
// 		// if err != nil {
// 		// 	slog.Error(" - error fetching from goodreads", "error", err)
// 		// } else {
// 		// 	if grb.SeriesName != "" {
// 		// 		slog.Info(" - found series info", "series", grb.SeriesName)
// 		// 	}
// 		// }
// 	} else {
// 		slog.Info(" - no goodreads id")
// 	}
// }
