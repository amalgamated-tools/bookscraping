package goodreads

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) GetSeries(seriesID string) (*Series, error) {
	url := fmt.Sprintf("%s/series/%s", c.baseURL, seriesID)
	fmt.Println("Fetching series from URL:", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching series: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing series HTML: %w", err)
	}

	return c.parseSeries(doc, url)
}

// parseSeries extracts series information from a goquery document
func (c *Client) parseSeries(doc *goquery.Document, url string) (*Series, error) {
	series := &Series{
		URL: url,
	}
	// Extract series ID from URL
	re := regexp.MustCompile(`/series/(\d+)`)
	if matches := re.FindStringSubmatch(url); len(matches) > 1 {
		series.ID = matches[1]
	}
	// Series name
	// series.Name = strings.TrimSpace(doc.Find("h1").First().Text())
	series.Title = strings.TrimSpace(doc.Find("div.responsiveSeriesHeader__title > h1").Text()) // Series name	// Description
	series.Works = strings.TrimSpace(doc.Find("div.responsiveSeriesHeader__subtitle.u-paddingBottomSmall").Text())

	series.Description = strings.TrimSpace(doc.Find("div.responsiveSeriesDescription > div.expandableHtml > span").Text())

	doc.Find("div.gr-col-md-8 > div.gr-boxBottomDivider > div > div.listWithDividers > div.listWithDividers__item").Each(func(i int, s *goquery.Selection) {
		bookNumber := strings.TrimSpace(s.Find("h3.gr-h3.gr-h3--noBottomMargin").Text())
		cover, _ := s.Find("div.responsiveBook > div.objectLockupContent > div.objectLockupContent__media > div.responsiveBook__media > a > img").Attr("src")
		title := strings.TrimSpace(s.Find("div.responsiveBook > div.objectLockupContent > div.u-paddingBottomXSmall > a > span[itemprop = 'name']").Text())
		bookURL, _ := s.Find("div.responsiveBook > div.objectLockupContent > div.u-paddingBottomXSmall > a").Attr("href")
		author := strings.TrimSpace(s.Find("div.responsiveBook > div.objectLockupContent > div.u-paddingBottomXSmall > div.u-paddingBottomTiny > span[itemprop = 'author'] > span[itemprop = 'name'] > a").Text())
		authorURL, _ := s.Find("div.responsiveBook > div.objectLockupContent > div.u-paddingBottomXSmall > div.u-paddingBottomTiny > span[itemprop = 'author'] > span[itemprop = 'name'] > a").Attr("href")
		rating := strings.TrimSpace(s.Find("div.responsiveBook > div.objectLockupContent > div.u-paddingBottomXSmall > div.communityRating > span").Text())

		fmt.Printf("Book %d:\n", i+1)
		fmt.Printf("  Book Number: %s\n", bookNumber)
		fmt.Printf("  Cover: %s\n", cover)
		fmt.Printf("  Title: %s\n", title)
		fmt.Printf("  Book URL: %s\n", bookURL)
		fmt.Printf("  Author: %s\n", author)
		fmt.Printf("  Author URL: %s\n", authorURL)
		fmt.Printf("  Rating: %s\n", rating)
		fmt.Println()
		book := SeriesBook{
			BookNumber: bookNumber,
			Book: Book{
				Title:         title,
				URL:           bookURL,
				CoverImageURL: cover,
			},
			// 	ID:         id,
			// 	Cover:      cover,
			// 	Title:      title,
			// 	BookURL:    bookURL,
			// 	Author:     author,
			// 	AuthorURL:  authorURL,
			// 	Rating:     rating,
		}
		book.Authors = append(book.Authors, Author{Name: author, URL: authorURL})
		series.Books = append(series.Books, book)
	})
	return series, nil
}
