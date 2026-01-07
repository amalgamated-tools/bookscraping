package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
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
	client := booklore.NewClient(bookloreServer, bookloreUsername, booklorePassword)
	client.Login()
	books, err := client.LoadAllBooks()
	if err != nil {
		panic(err)
	}
	grClient := goodreads.NewClient()
	series := make(map[string]map[float64]booklore.Book)
	for _, book := range books {
		if book.GoodreadsId != "" {
			b, err := grClient.GetBook(book.GoodreadsId)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Fetched book info from Goodreads: %s\n", b.Title)
		}
		if book.SeriesName != "" {
			if _, ok := series[book.SeriesName]; !ok {
				series[book.SeriesName] = make(map[float64]booklore.Book)
			}
			series[book.SeriesName][book.SeriesNumber] = book
		}
	}
	jsonBooks, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("books.json", jsonBooks, 0o644)
}
