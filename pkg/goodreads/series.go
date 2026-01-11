package goodreads

import (
	"fmt"
	"log/slog"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) GetSeries(seriesID string) (*Series, error) {
	url := fmt.Sprintf("%s/series/%s", c.baseURL, seriesID)
	fmt.Println("Fetching series from URL:", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching series page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing series HTML: %w", err)
	}

	doc.Find("div[data-react-class='ReactComponents.SeriesList']").Each(func(i int, s *goquery.Selection) {
		fmt.Println("Found React series div", i)
		dataProps, exists := s.Attr("data-react-props")
		slog.Debug("dataProps", slog.String("dataProps", dataProps), slog.Bool("exists", exists))
		b, err := ParseSeriesData([]byte(dataProps))
		if err != nil {
			slog.Error("Error parsing series data", slog.Int("index", i), slog.String("error", err.Error()))
			return
		}
		for _, bb := range b {
			slog.Debug("Parsed book in series", slog.Int("index", i), slog.String("bookTitle", bb.Book.Title), slog.String("seriesPosition", bb.SeriesPosition))
		}
	})

	return nil, nil
	// return c.parseReactSeries(doc, url)
}

// // parseReactSeries extracts series information from a React-based Goodreads series page
// func (c *Client) parseReactSeries(doc *goquery.Document, url string) (*Series, error) {
// 	series := &Series{
// 		URL: url,
// 	}
// 	blah := make(map[string]SeriesBookk)
// 	doc.Find("div[data-react-class='ReactComponents.SeriesList']").Each(func(i int, s *goquery.Selection) {
// 		slog.Debug("Found React series div", slog.Int("index", i))
// 		dataProps, exists := s.Attr("data-react-props")
// 		if exists {
// 			slog.Debug("Found data-react-props", slog.Int("index", i), slog.String("dataProps", dataProps))
// 			booksWithPosition, err := ParseSeriesData([]byte(dataProps))
// 			if err != nil {
// 				slog.Error("Error parsing series data", slog.Int("index", i), slog.String("error", err.Error()))
// 				return
// 			}
// 			for _, bwp := range booksWithPosition {
// 				blah[bwp.SeriesPosition] = bwp.Book
// 			}
// 		} else {
// 			slog.Warn("No data-react-props found", slog.Int("index", i))
// 		}
// 	})

// 	return series, nil
// }
