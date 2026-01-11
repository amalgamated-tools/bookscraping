package goodreads

import "encoding/json"

// SeriesData represents the JSON structure containing series information
type SeriesData struct {
	Series        []SeriesEntry `json:"series"`
	SeriesHeaders []string      `json:"seriesHeaders"`
}

// SeriesEntry represents a single entry in the series array
type SeriesEntry struct {
	IsLibrarianView bool        `json:"isLibrarianView"`
	Book            SeriesBookk `json:"book"`
}

// SeriesBookk represents book data within a series
type SeriesBookk struct {
	ImageURL         string            `json:"imageUrl"`
	BookID           string            `json:"bookId"`
	WorkID           string            `json:"workId"`
	BookURL          string            `json:"bookUrl"`
	FromSearch       bool              `json:"from_search"`
	FromSRP          bool              `json:"from_srp"`
	QID              interface{}       `json:"qid"`
	Rank             interface{}       `json:"rank"`
	Title            string            `json:"title"`
	BookTitleBare    string            `json:"bookTitleBare"`
	NumPages         int               `json:"numPages"`
	AvgRating        float64           `json:"avgRating"`
	RatingsCount     int               `json:"ratingsCount"`
	Author           SeriesAuthor      `json:"author"`
	KCRPreviewURL    string            `json:"kcrPreviewUrl"`
	Description      SeriesDescription `json:"description"`
	TextReviewsCount int               `json:"textReviewsCount"`
	PublicationDate  string            `json:"publicationDate"`
	ToBePublished    bool              `json:"toBePublished"`
	Editions         string            `json:"editions"`
	EditionsURL      string            `json:"editionsUrl"`
}

// SeriesAuthor represents author information
type SeriesAuthor struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	IsGoodreadsAuthor bool   `json:"isGoodreadsAuthor"`
	ProfileURL        string `json:"profileUrl"`
	WorksListURL      string `json:"worksListUrl"`
}

// SeriesDescription represents book description
type SeriesDescription struct {
	TruncatedHTML string `json:"truncatedHtml"`
	HTML          string `json:"html"`
}

// BookWithPosition combines a book with its position in the series
type BookWithPosition struct {
	Book           SeriesBookk
	SeriesPosition string
	Index          int
}

// ParseSeriesData parses the JSON data and returns books with their series positions
func ParseSeriesData(jsonData []byte) ([]BookWithPosition, error) {
	var data SeriesData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	result := make([]BookWithPosition, 0, len(data.Series))
	for i, entry := range data.Series {
		position := ""
		if i < len(data.SeriesHeaders) {
			position = data.SeriesHeaders[i]
		}

		result = append(result, BookWithPosition{
			Book:           entry.Book,
			SeriesPosition: position,
			Index:          i,
		})
	}

	return result, nil
}
