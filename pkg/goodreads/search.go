package bookscraping

// // Search performs a search on Goodreads
// func (c *Client) Search(query string, page int) (*SearchResult, error) {
// 	searchURL := fmt.Sprintf("%s/search?q=%s&page=%d", c.baseURL, url.QueryEscape(query), page)
// 	doc, err := c.parseHTML(searchURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("performing search: %w", err)
// 	}

// 	result := &SearchResult{}

// 	// Parse book results
// 	doc.Find("tr[itemtype='http://schema.org/Book']").Each(func(i int, s *goquery.Selection) {
// 		book := Book{}

// 		// Title and URL
// 		titleLink := s.Find("a.bookTitle")
// 		book.Title = strings.TrimSpace(titleLink.Text())
// 		bookURL, _ := titleLink.Attr("href")
// 		if bookURL != "" {
// 			book.URL = c.baseURL + bookURL

// 			// Extract book ID
// 			re := regexp.MustCompile(`/book/show/(\d+)`)
// 			if matches := re.FindStringSubmatch(bookURL); len(matches) > 1 {
// 				book.ID = matches[1]
// 			}
// 		}

// 		// Author
// 		authorLink := s.Find("a.authorName")
// 		if authorLink.Length() > 0 {
// 			authorName := strings.TrimSpace(authorLink.Text())
// 			authorURL, _ := authorLink.Attr("href")

// 			author := Author{
// 				Name: authorName,
// 			}

// 			if authorURL != "" {
// 				author.URL = c.baseURL + authorURL

// 				// Extract author ID
// 				re := regexp.MustCompile(`/author/show/(\d+)`)
// 				if matches := re.FindStringSubmatch(authorURL); len(matches) > 1 {
// 					author.ID = matches[1]
// 				}
// 			}

// 			book.Authors = append(book.Authors, author)
// 		}

// 		// Cover image
// 		imgSrc, _ := s.Find("img.bookCover").Attr("src")
// 		book.CoverImageURL = imgSrc

// 		// Rating
// 		ratingStr := s.Find("span.minirating").Text()
// 		book.Rating = parseFloat(ratingStr)

// 		// Extract rating and review counts
// 		re := regexp.MustCompile(`([\d,]+)\s+ratings`)
// 		if matches := re.FindStringSubmatch(ratingStr); len(matches) > 1 {
// 			book.RatingCount = parseCount(matches[1])
// 		}

// 		re = regexp.MustCompile(`([\d,]+)\s+reviews`)
// 		if matches := re.FindStringSubmatch(ratingStr); len(matches) > 1 {
// 			book.ReviewCount = parseCount(matches[1])
// 		}

// 		// Published year
// 		pubText := s.Find("span.greyText.smallText").Text()
// 		re = regexp.MustCompile(`published\s+(\d{4})`)
// 		if matches := re.FindStringSubmatch(pubText); len(matches) > 1 {
// 			book.PublishedYear = matches[1]
// 		}

// 		if book.Title != "" {
// 			result.Books = append(result.Books, book)
// 		}
// 	})

// 	result.TotalResults = len(result.Books)

// 	return result, nil
// }

// // SearchBooks is a convenience method for searching only books
// func (c *Client) SearchBooks(query string) ([]Book, error) {
// 	result, err := c.Search(query, 1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result.Books, nil
// }
