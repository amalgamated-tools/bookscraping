package main

import (
	_ "embed"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	_ "modernc.org/sqlite"
)

func main() {
	bookloreServer := os.Getenv("BOOKLORE_SERVER")
	if bookloreServer == "" {
		panic("BOOKLORE_SERVER environment variable is not set")
	}
	bookloreUsername := os.Getenv("BOOKLORE_USERNAME")
	if bookloreUsername == "" {
		panic("BOOKLORE_USERNAME environment variable is not set")
	}
	booklorePassword := os.Getenv("BOOKLORE_PASSWORD")
	if booklorePassword == "" {
		panic("BOOKLORE_PASSWORD environment variable is not set")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	client := booklore.NewClient(bookloreServer, bookloreUsername, booklorePassword)
	client.Login()

	books, err := client.LoadAllBooks()
	if err != nil {
		panic(err)
	}

	series := make(map[string]map[float64]booklore.Book)
	for _, book := range books {
		slog.Info("Book loaded", "title", book.Title, "series", book.SeriesName, "number", book.SeriesNumber)
		if book.SeriesName != "" {
			if _, ok := series[book.SeriesName]; !ok {
				series[book.SeriesName] = make(map[float64]booklore.Book)
			}
			series[book.SeriesName][book.SeriesNumber] = book
		} else {
			slog.Info("Book without series", "title", book.Title)
		}
	}

}
