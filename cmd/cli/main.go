package main

import (
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/config"
	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// grClient := goodreads.NewClient()

	client := booklore.NewClient(cfg.BookloreServer, cfg.BookloreUsername, cfg.BooklorePassword)
	client.Login()

	books, err := client.LoadAllBooks()
	if err != nil {
		panic(err)
	}

	// series := make(map[string]map[float64]booklore.Book)
	for _, book := range books {
		// check if we have a goodreads id
		if book.GoodreadsId == "" {
			slog.Info("No goodreads id for book", slog.String("title", book.Title), slog.Int64("id", book.ID))
			slog.Info(fmt.Sprintf("https://booklore.veverka.net/book/%d", book.ID))
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
