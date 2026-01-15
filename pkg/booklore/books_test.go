package booklore

import (
	"reflect"
	"testing"
)

func TestProcessBookJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Book
	}{
		{
			name: "Full book data",
			jsonData: `{
    "id": 27,
    "bookType": "EPUB",
    "libraryId": 1,
    "libraryName": "My Library",
    "fileName": "Die Twice - Andrew Grant.epub",
    "fileSubPath": "Andrew Grant/Die Twice (24)",
    "fileSizeKb": 1067,
    "addedOn": "2025-11-14T00:16:36Z",
    "metadata": {
      "bookId": 27,
      "title": "Die Twice",
      "publisher": "A Thomas Dunne Book",
      "publishedDate": "2010-01-01",
      "seriesName": "David Trevellyan",
      "seriesNumber": 2,
      "seriesTotal": 3,
      "isbn13": "9780230747586",
      "isbn10": "0230747582",
      "pageCount": 306,
      "language": "English",
      "goodreadsId": "7315139-die-twice",
      "goodreadsRating": 3.42,
      "goodreadsReviewCount": 714,
      "hardcoverId": "die-twice",
      "hardcoverRating": 3,
      "hardcoverReviewCount": 4,
      "coverUpdatedOn": "2025-11-20T12:16:55Z",
      "authors": [
        "Andrew Grant"
      ],
      "categories": [
        "War",
        "Fiction"
      ],
      "moods": [],
      "tags": [],
      "titleLocked": false,
      "subtitleLocked": false,
      "publisherLocked": false,
      "publishedDateLocked": false,
      "descriptionLocked": false,
      "seriesNameLocked": false,
      "seriesNumberLocked": false,
      "seriesTotalLocked": false,
      "isbn13Locked": false,
      "isbn10Locked": false,
      "asinLocked": false,
      "goodreadsIdLocked": false,
      "comicvineIdLocked": false,
      "hardcoverIdLocked": false,
      "hardcoverBookIdLocked": false,
      "googleIdLocked": false,
      "pageCountLocked": false,
      "languageLocked": false,
      "amazonRatingLocked": false,
      "amazonReviewCountLocked": false,
      "goodreadsRatingLocked": false,
      "goodreadsReviewCountLocked": false,
      "hardcoverRatingLocked": false,
      "hardcoverReviewCountLocked": false,
      "lubimyczytacIdLocked": false,
      "lubimyczytacRatingLocked": false,
      "coverLocked": false,
      "authorsLocked": false,
      "categoriesLocked": false,
      "moodsLocked": false,
      "tagsLocked": false
    },
    "metadataMatchScore": 82.0225,
    "shelves": [],
    "libraryPath": {
      "id": 1
    }
  }`,
			expected: Book{
				ID:              27,
				Title:           "Die Twice",
				Description:     "",
				SeriesName:      "David Trevellyan",
				SeriesNumber:    2.0,
				SeriesTotal:     3,
				ISBN13:          "9780230747586",
				ISBN10:          "0230747582",
				ASIN:            "",
				HardCoverID:     "die-twice",
				HardCoverBookID: 0,
				GoodreadsId:     "7315139-die-twice",
				GoogleId:        "",
				Authors:         []string{"Andrew Grant"},
			},
		},
		{
			name: "SeriesID as string",
			jsonData: `{
				"id": 2,
				"metadata": {
					"title": "Book with String SeriesID",
					"seriesId": "200"
				}
			}`,
			expected: Book{
				ID:      2,
				Title:   "Book with String SeriesID",
				Authors: []string{},
			},
		},
		{
			name: "Multiple authors",
			jsonData: `{
				"id": 3,
				"metadata": {
					"title": "Collaboration",
					"authors": ["Author One", "Author Two"]
				}
			}`,
			expected: Book{
				ID:      3,
				Title:   "Collaboration",
				Authors: []string{"Author One", "Author Two"},
			},
		},
		{
			name: "Missing metadata",
			jsonData: `{
				"id": 4
			}`,
			expected: Book{
				ID:      4,
				Authors: []string{},
			},
		},
		{
			name: "Empty authors",
			jsonData: `{
				"id": 5,
				"metadata": {
					"title": "No Author Book",
					"authors": []
				}
			}`,
			expected: Book{
				ID:      5,
				Title:   "No Author Book",
				Authors: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processBookJSON([]byte(tt.jsonData))

			// Compare slices manually if needed, or use DeepEqual
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("processBookJSON() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}
