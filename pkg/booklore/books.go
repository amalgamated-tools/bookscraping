package booklore

import (
	"io"
	"net/http"

	"github.com/buger/jsonparser"
)

func (c *Client) LoadAllBooks() ([]Book, error) {
	url := c.baseURL + "/api/v1/books?withDescription=true"
	books := []Book{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "*/*")
	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	// read the response body
	// #nosec G304
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		books = append(books, processBookJSON(value))
	})

	return books, nil
}

func processBookJSON(value []byte) Book {
	book := Book{}
	book.ID, _ = jsonparser.GetInt(value, "id")

	book.Title, _ = jsonparser.GetString(value, "metadata", "title")
	book.Description, _ = jsonparser.GetString(value, "metadata", "description")
	book.SeriesName, _ = jsonparser.GetString(value, "metadata", "seriesName")
	book.SeriesNumber, _ = jsonparser.GetFloat(value, "metadata", "seriesNumber")
	book.SeriesTotal, _ = jsonparser.GetInt(value, "metadata", "seriesTotal")
	book.ISBN13, _ = jsonparser.GetString(value, "metadata", "isbn13")
	book.ISBN10, _ = jsonparser.GetString(value, "metadata", "isbn10")
	book.ASIN, _ = jsonparser.GetString(value, "metadata", "asin")
	book.HardCoverID, _ = jsonparser.GetString(value, "metadata", "hardcoverId")
	book.HardCoverBookID, _ = jsonparser.GetInt(value, "metadata", "hardcoverBookId")
	book.GoodreadsId, _ = jsonparser.GetString(value, "metadata", "goodreadsId")
	book.GoogleId, _ = jsonparser.GetString(value, "metadata", "googleId")
	authors := []string{}
	jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		author, _ := jsonparser.ParseString(value)
		authors = append(authors, author)
	}, "metadata", "authors")
	book.Authors = authors

	return book
}

type Book struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	SeriesName      string   `json:"seriesName"`
	SeriesNumber    float64  `json:"seriesNumber"`
	SeriesTotal     int64    `json:"seriesTotal"`
	ISBN13          string   `json:"isbn13"`
	ISBN10          string   `json:"isbn10"`
	ASIN            string   `json:"asin"`
	HardCoverID     string   `json:"hardcoverId"`
	HardCoverBookID int64    `json:"hardcoverBookId"`
	Authors         []string `json:"authors"`
	GoodreadsId     string   `json:"goodreadsId"`
	GoogleId        string   `json:"googleId"`
}
