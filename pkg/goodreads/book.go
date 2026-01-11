package goodreads

// // GetBook fetches and parses a book by its ID
// func (c *Client) GetBook(bookID string) (*Book, error) {
// 	url := fmt.Sprintf("%s/book/show/%s", c.baseURL, bookID)

// 	resp, err := c.httpClient.Get(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("fetching book: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	doc, err := goquery.NewDocumentFromReader(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return c.parseBook(doc, url)
// }

// func (c *Client) GetSeriesByBookID(bookID string) (*Series, error) {
// 	url := fmt.Sprintf("%s/book/show/%s", c.baseURL, bookID)

// 	resp, err := c.httpClient.Get(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("fetching book: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	doc, err := goquery.NewDocumentFromReader(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	book := &Book{
// 		BookURL: url,
// 	}

// 	book.parseSeries(doc)
// 	return c.GetSeries(book.SeriesID)
// }

// // // parseBook extracts book information from a goquery document
// func (c *Client) parseBook(doc *goquery.Document, url string) (*Book, error) {
// 	book := &Book{
// 		BookURL: url,
// 	}

// 	// Extract book ID from URL
// 	re := regexp.MustCompile(`/book/show/(\d+)`)
// 	if matches := re.FindStringSubmatch(url); len(matches) > 1 {
// 		book.BookID = matches[1]
// 	}

// 	// Title
// 	book.parseTitle(doc)

// 	// 	// Authors
// 	book.parseAuthors(doc)

// 	// Rating
// 	book.parseRating(doc)

// 	// Description
// 	book.parseDescription(doc)

// 	// Cover image
// 	book.parseImage(doc)

// 	// Page count
// 	book.parsePageCount(doc)

// 	// Publication info
// 	book.parsePublicationDetails(doc)

// 	// ISBN
// 	book.parseISBN(doc)

// 	// Genres
// 	book.parseGenres(doc)
// 	// Series information
// 	book.parseSeries(doc)

// 	return book, nil
// }

// func (book *Book) parsePublicationDetails(doc *goquery.Document) {
// 	pubInfo := doc.Find("p[data-testid='publicationInfo']").Text()
// 	if pubInfo != "" {
// 		// Extract year
// 		re := regexp.MustCompile(`(\d{4})`)
// 		if matches := re.FindStringSubmatch(pubInfo); len(matches) > 0 {
// 			book.PublishedYear = matches[0]
// 		}

// 		// Extract publisher
// 		parts := strings.Split(pubInfo, "by")
// 		if len(parts) > 1 {
// 			book.Publisher = strings.TrimSpace(parts[1])
// 		}
// 	}
// }

// func (book *Book) parseISBN(doc *goquery.Document) {
// 	doc.Find("div[class*='BookDetails'] div[class*='TruncatedContent']").Each(func(i int, s *goquery.Selection) {
// 		text := s.Text()
// 		if strings.Contains(text, "ISBN") {
// 			re := regexp.MustCompile(`ISBN[:\s]*(\d+)`)
// 			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
// 				if len(matches[1]) == 13 {
// 					book.ISBN13 = matches[1]
// 				} else {
// 					book.ISBN = matches[1]
// 				}
// 			}
// 		}
// 	})
// }

// func (book *Book) parseGenres(doc *goquery.Document) {
// 	doc.Find("span[class*='BookPageMetadataSection__genreButton'] a").Each(func(i int, s *goquery.Selection) {
// 		genre := strings.TrimSpace(s.Text())
// 		if genre != "" {
// 			book.Genres = append(book.Genres, genre)
// 		}
// 	})

// }

// func (book *Book) parsePageCount(doc *goquery.Document) {
// 	doc.Find("p[data-testid='pagesFormat']").Each(func(i int, s *goquery.Selection) {
// 		text := s.Text()
// 		re := regexp.MustCompile(`(\d+)\s+pages`)
// 		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
// 			book.PageCount, _ = strconv.Atoi(matches[1])
// 		}
// 	})
// }

// func (book *Book) parseImage(doc *goquery.Document) {
// 	imgSrc, _ := doc.Find("img[class*='ResponsiveImage']").Attr("src")
// 	if imgSrc == "" {
// 		imgSrc, _ = doc.Find(".BookPage__bookCover img").Attr("src")
// 	}
// 	book.ImageURL = imgSrc
// }

// func (book *Book) parseDescription(doc *goquery.Document) {
// 	desc := doc.Find("div[data-testid='description'] span[class*='Formatted']").First()
// 	book.Description = strings.TrimSpace(desc.Text())
// 	if book.Description == "" {
// 		book.Description = strings.TrimSpace(doc.Find(".DetailsLayoutRightParagraph__widthConstrained").Text())
// 	}
// }

// func (book *Book) parseSeries(doc *goquery.Document) {
// 	seriesLink := doc.Find("h3[class*='Text__italic'] a").First()
// 	if seriesLink.Length() > 0 {
// 		seriesText := strings.TrimSpace(seriesLink.Text())
// 		seriesURL, _ := seriesLink.Attr("href")

// 		book.SeriesName = seriesText
// 		book.SeriesURL = seriesURL

// 		// extract series ID from URL
// 		re := regexp.MustCompile(`/series/(\d+)`)
// 		if matches := re.FindStringSubmatch(seriesURL); len(matches) > 1 {
// 			book.SeriesID = matches[1]
// 		}

// 		// Extract series position from surrounding text
// 		seriesParent := seriesLink.Parent().Text()
// 		ore := regexp.MustCompile(`#([\d.]+)`)
// 		if matches := ore.FindStringSubmatch(seriesParent); len(matches) > 1 {
// 			book.SeriesPosition = matches[1]
// 		}

// 		// Store full series info if needed
// 		if seriesURL != "" {
// 			book.SeriesName = seriesText
// 		}
// 	}
// }

// func (book *Book) parseRating(doc *goquery.Document) {
// 	ratingDiv := doc.Find("div[class*='RatingStatistics__rating']")
// 	if ratingDiv.Length() > 1 {
// 		slog.Debug("Multiple rating divs found, using the first one")
// 		ratingDiv = ratingDiv.First()
// 	}
// 	ratingStr := ratingDiv.Text()
// 	if ratingStr == "" {
// 		slog.Debug("Rating not found in primary selector, trying fallback")
// 		ratingStr = doc.Find(".RatingStatistics__rating").Text()
// 	}

// 	book.Rating = parseFloat(ratingStr)
// }

// func (book *Book) parseAuthors(doc *goquery.Document) {
// 	doc.Find("span[data-testid='name']").Each(func(i int, s *goquery.Selection) {
// 		authorName := strings.TrimSpace(s.Text())
// 		if authorName != "" {
// 			authorLink := s.Parent()
// 			authorURL, _ := authorLink.Attr("href")

// 			author := Author{
// 				Name: authorName,
// 			}

// 			if authorURL != "" {
// 				author.URL = authorURL
// 				// Extract author ID from URL
// 				re := regexp.MustCompile(`/author/show/(\d+)`)
// 				if matches := re.FindStringSubmatch(authorURL); len(matches) > 1 {
// 					author.ID = matches[1]
// 				}
// 			}

// 			book.Authors = append(book.Authors, author)
// 		}
// 	})

// 	// Fallback for authors
// 	if len(book.Authors) == 0 {
// 		doc.Find(".ContributorLink__name").Each(func(i int, s *goquery.Selection) {
// 			authorName := strings.TrimSpace(s.Text())
// 			if authorName != "" {
// 				book.Authors = append(book.Authors, Author{Name: authorName})
// 			}
// 		})
// 	}
// }

// func (book *Book) parseTitle(doc *goquery.Document) {
// 	book.Title = strings.TrimSpace(doc.Find("h1[data-testid='bookTitle']").Text())
// 	if book.Title == "" {
// 		book.Title = strings.TrimSpace(doc.Find(".BookPageTitleSection__title h1").Text())
// 	}
// 	if book.Title == "" {
// 		book.Title = strings.TrimSpace(doc.Find("h1.Text__title1").Text())
// 	}
// }

// // parseFloat extracts a float from a string
// func parseFloat(s string) float64 {
// 	s = strings.TrimSpace(s)
// 	s = regexp.MustCompile(`[^\d.]`).ReplaceAllString(s, "")
// 	f, _ := strconv.ParseFloat(s, 64)
// 	return f
// }

// // // parseCount extracts an integer count from a string (handles "1,234" format)
// // func parseCount(s string) int {
// // 	s = strings.TrimSpace(s)
// // 	// Remove non-digit characters except for digits
// // 	re := regexp.MustCompile(`[\d,]+`)
// // 	match := re.FindString(s)
// // 	if match == "" {
// // 		return 0
// // 	}

// // 	// Remove commas
// // 	match = strings.ReplaceAll(match, ",", "")
// // 	count, _ := strconv.Atoi(match)
// // 	return count
// // }
