package bookscraping

// // GetAuthor fetches and parses an author by their ID
// func (c *Client) GetAuthor(authorID string) (*Author, error) {
// 	url := fmt.Sprintf("%s/author/show/%s", c.baseURL, authorID)
// 	doc, err := c.parseHTML(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("fetching author: %w", err)
// 	}

// 	author := &Author{
// 		ID:  authorID,
// 		URL: url,
// 	}

// 	// Name
// 	author.Name = strings.TrimSpace(doc.Find("h1[class*='AuthorName']").Text())
// 	if author.Name == "" {
// 		author.Name = strings.TrimSpace(doc.Find(".authorName span").Text())
// 	}

// 	// Image
// 	imgSrc, _ := doc.Find("img[class*='AuthorImage']").Attr("src")
// 	if imgSrc == "" {
// 		imgSrc, _ = doc.Find(".leftContainer img").Attr("src")
// 	}
// 	author.ImageURL = imgSrc

// 	// Bio
// 	bioDiv := doc.Find("div[class*='aboutAuthorInfo']").First()
// 	if bioDiv.Length() == 0 {
// 		bioDiv = doc.Find(".aboutAuthorInfo").First()
// 	}
// 	author.Bio = strings.TrimSpace(bioDiv.Text())

// 	// Born/Died dates
// 	doc.Find("div[class*='dataItem']").Each(func(i int, s *goquery.Selection) {
// 		label := strings.TrimSpace(s.Find("span[class*='dataTitle']").Text())
// 		value := strings.TrimSpace(s.Find("span[class*='dataValue']").Text())

// 		switch {
// 		case strings.Contains(strings.ToLower(label), "born"):
// 			author.BornAt = value
// 		case strings.Contains(strings.ToLower(label), "died"):
// 			author.DiedAt = value
// 		case strings.Contains(strings.ToLower(label), "website"):
// 			author.Website = value
// 		}
// 	})

// 	// Stats
// 	doc.Find("div[class*='hreview-aggregate']").Each(func(i int, s *goquery.Selection) {
// 		text := s.Text()

// 		// Average rating
// 		re := regexp.MustCompile(`([\d.]+)\s+avg\s+rating`)
// 		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
// 			author.AverageRating = parseFloat(matches[1])
// 		}

// 		// Rating count
// 		re = regexp.MustCompile(`([\d,]+)\s+ratings`)
// 		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
// 			author.RatingCount = parseCount(matches[1])
// 		}

// 		// Review count
// 		re = regexp.MustCompile(`([\d,]+)\s+reviews`)
// 		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
// 			author.ReviewCount = parseCount(matches[1])
// 		}
// 	})

// 	// Genres
// 	doc.Find("div[class*='dataItem'] a[href*='/genres/']").Each(func(i int, s *goquery.Selection) {
// 		genre := strings.TrimSpace(s.Text())
// 		if genre != "" {
// 			author.Genres = append(author.Genres, genre)
// 		}
// 	})

// 	// Influenced by
// 	doc.Find("div[class*='dataItem']:contains('Influenced by') a").Each(func(i int, s *goquery.Selection) {
// 		influence := strings.TrimSpace(s.Text())
// 		if influence != "" {
// 			author.InfluencedBy = append(author.InfluencedBy, influence)
// 		}
// 	})

// 	// Related authors
// 	doc.Find("div[class*='similarAuthor'] a").Each(func(i int, s *goquery.Selection) {
// 		relatedAuthor := strings.TrimSpace(s.Text())
// 		if relatedAuthor != "" {
// 			author.RelatedAuthors = append(author.RelatedAuthors, relatedAuthor)
// 		}
// 	})

// 	// Fans count
// 	fansText := doc.Find("span[class*='fansCount']").Text()
// 	author.FansCount = parseCount(fansText)

// 	return author, nil
// }

// // GetAuthorBooks fetches books by an author
// func (c *Client) GetAuthorBooks(authorID string, page int) ([]Book, error) {
// 	url := fmt.Sprintf("%s/author/list/%s?page=%d", c.baseURL, authorID, page)
// 	doc, err := c.parseHTML(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("fetching author books: %w", err)
// 	}

// 	var books []Book

// 	doc.Find("tr[itemtype='http://schema.org/Book']").Each(func(i int, s *goquery.Selection) {
// 		book := Book{}

// 		// Title and URL
// 		titleLink := s.Find("a[class*='bookTitle']")
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

// 		// Cover image
// 		imgSrc, _ := s.Find("img[class*='bookCover']").Attr("src")
// 		book.CoverImageURL = imgSrc

// 		// Rating
// 		ratingStr := s.Find("span[class*='minirating']").Text()
// 		book.Rating = parseFloat(ratingStr)

// 		// Published year
// 		pubText := s.Find("span[class*='greyText']").Text()
// 		re := regexp.MustCompile(`published\s+(\d{4})`)
// 		if matches := re.FindStringSubmatch(pubText); len(matches) > 1 {
// 			book.PublishedYear = matches[1]
// 		}

// 		if book.Title != "" {
// 			books = append(books, book)
// 		}
// 	})

// 	return books, nil
// }
