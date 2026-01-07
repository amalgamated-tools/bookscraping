# BiblioReads Go Library

A Go library for scraping book data from Goodreads. This is a Go alternative to the [BiblioReads](https://github.com/nesaku/BiblioReads) project, built using **goquery** (the Go equivalent of Cheerio).

## Features

- 📚 Fetch book details (title, author, rating, description, ISBN, genres, etc.)
- ✍️ Get author information and their books
- 🔍 Search for books and authors
- 💬 Retrieve quotes from books
- 📖 Get series information
- 📝 Fetch Goodreads lists
- 🚀 No API key required (web scraping)
- 🔒 Privacy-focused (proxy your requests)

## Installation

```bash
go get github.com/amalgamated-tools/bookscraping
```

## Dependencies

This library uses [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing:

```bash
go get github.com/PuerkitoBio/goquery
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/amalgamated-tools/bookscraping"
)

func main() {
    // Create a new client
    client := bookscraping.NewClient()
    
    // Get a book by ID
    book, err := client.GetBook("5907") // The Hobbit
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Title: %s\n", book.Title)
    fmt.Printf("Author: %s\n", book.Authors[0].Name)
    fmt.Printf("Rating: %.2f\n", book.Rating)
}
```

## Usage Examples

### Get Book Details

```go
// By book ID
book, err := client.GetBook("5907")

// By full URL
book, err := client.GetBookByURL("https://www.goodreads.com/book/show/5907.The_Hobbit")
```

### Search for Books

```go
// Search and get all results
results, err := client.Search("Harry Potter", 1)

// Or use convenience method for books only
books, err := client.SearchBooks("Harry Potter")

for _, book := range books {
    fmt.Printf("%s by %s\n", book.Title, book.Authors[0].Name)
}
```

### Get Author Information

```go
author, err := client.GetAuthor("656983") // J.R.R. Tolkien

fmt.Printf("Name: %s\n", author.Name)
fmt.Printf("Born: %s\n", author.BornAt)
fmt.Printf("Bio: %s\n", author.Bio)
fmt.Printf("Average Rating: %.2f\n", author.AverageRating)
```

### Get Author's Books

```go
books, err := client.GetAuthorBooks("656983", 1) // page 1

for _, book := range books {
    fmt.Printf("%s (%.2f stars)\n", book.Title, book.Rating)
}
```

### Get Book Quotes

```go
quotes, err := client.GetQuotes("5907", 1) // page 1

for _, quote := range quotes {
    fmt.Printf("Quote: %s\n", quote.Text)
    fmt.Printf("Likes: %d\n", quote.Likes)
}
```

### Get Series Information

```go
series, err := client.GetSeries("66") // The Lord of the Rings

fmt.Printf("Series: %s\n", series.Name)
fmt.Printf("Books: %d\n", series.BookCount)

for _, seriesBook := range series.Books {
    fmt.Printf("Book %s: %s\n", seriesBook.Position, seriesBook.Book.Title)
}
```

### Get Goodreads Lists

```go
list, err := client.GetList("1.Best_Books_Ever")

fmt.Printf("List: %s\n", list.Title)
fmt.Printf("Books: %d\n", list.BookCount)
```

## Configuration

### Custom Timeout

```go
client := bookscraping.NewClient().WithTimeout(60 * time.Second)
```

### Custom User Agent

```go
client := bookscraping.NewClient().WithUserAgent("MyApp/1.0")
```

## Data Structures

### Book

```go
type Book struct {
    ID             string
    Title          string
    Authors        []Author
    ISBN           string
    ISBN13         string
    Description    string
    Rating         float64
    RatingCount    int
    ReviewCount    int
    PageCount      int
    PublishedYear  string
    Publisher      string
    Language       string
    CoverImageURL  string
    Genres         []string
    SeriesName     string
    SeriesPosition string
    URL            string
}
```

### Author

```go
type Author struct {
    ID              string
    Name            string
    URL             string
    ImageURL        string
    Bio             string
    BornAt          string
    DiedAt          string
    Website         string
    Genres          []string
    InfluencedBy    []string
    AverageRating   float64
    RatingCount     int
    ReviewCount     int
    FansCount       int
    RelatedAuthors  []string
}
```

## How It Works

This library scrapes Goodreads pages using goquery (jQuery-like selectors for Go). Since Goodreads deprecated their public API in 2020, web scraping is the only way to programmatically access their data.

**Note:** Web scraping may be affected by changes to Goodreads' HTML structure. Always respect Goodreads' terms of service and rate limits.

## Comparison with BiblioReads

| Feature | BiblioReads (Node.js) | This Library (Go) |
|---------|----------------------|-------------------|
| Language | JavaScript/TypeScript | Go |
| HTML Parser | Cheerio | goquery |
| Use Case | Web front-end | Backend/CLI/Library |
| Dependencies | Next.js, React | Minimal (just goquery) |
| Performance | Good | Excellent |

## Examples

See the [examples](examples/) directory for complete working examples.

## License

GNU AGPLv3 - See LICENSE for details

## Disclaimer

This library is not affiliated with Goodreads or Amazon. It scrapes publicly available data from Goodreads. Please use responsibly and respect Goodreads' terms of service.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Inspired By

- [BiblioReads](https://github.com/nesaku/BiblioReads) - Privacy-focused Goodreads front-end
- [goquery](https://github.com/PuerkitoBio/goquery) - jQuery-like HTML parsing for Go
