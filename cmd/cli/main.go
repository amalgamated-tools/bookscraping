package main

import (
	"encoding/json"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
)

func main() {
	client := booklore.NewClient("https://bl.veverka.net", "dev", "devdevdev")
	client.Login()
	books, err := client.LoadAllBooks()
	if err != nil {
		panic(err)
	}
	series := make(map[string]map[float64]booklore.Book)
	for _, book := range books {
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
