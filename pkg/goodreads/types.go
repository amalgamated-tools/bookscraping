package goodreads

// // Book represents a book from Goodreads
// type Book struct {
// 	BookID         string `json:"bookId"`
// 	WorkID         string `json:"workId"`
// 	Description    string
// 	ImageURL       string `json:"imageUrl"`
// 	Title          string
// 	BookURL        string `json:"bookUrl"`
// 	Authors        []Author
// 	Rating         float64
// 	PageCount      int
// 	PublishedYear  string
// 	Language       string
// 	Publisher      string
// 	ISBN           string
// 	ISBN13         string
// 	Genres         []string
// 	SeriesName     string
// 	SeriesPosition string
// 	SeriesURL      string
// 	SeriesID       string
// }

// type SeriesBook struct {
// 	BookNumber int64
// 	Book
// }

// type Author struct {
// 	Name string
// 	ID   string
// 	URL  string
// }

type Series struct {
	ID    string
	Title string
	// Description string
	// Works       string
	// WorksCount  int64
	// URL         string
	// Books       []SeriesBook
}

// // }	URL         string `json:"url"`	Books       []Book `json:"books,omitempty"`	VoterCount  int    `json:"voter_count"`	BookCount   int    `json:"book_count"`	Description string `json:"description,omitempty"`	Title       string `json:"title"`	ID          string `json:"id"`
// //
// // type List struct {// List represents a Goodreads list}	TotalResults int `json:"total_results"`	Authors []Author `json:"authors"`	Books   []Book `json:"books"`
// type SearchResult struct {
// 	Books []Book `json:"books,omitempty"`
// }

// //
// // type SeriesBook struct {// SeriesBook represents a book in a series}	URL         string      `json:"url"`	Books       []SeriesBook `json:"books,omitempty"`	BookCount   int         `json:"book_count"`	Description string      `json:"description,omitempty"`	Name        string      `json:"name"`	ID          string      `json:"id"`
// //
// // type Series struct {// Series represents a book series}	Likes    int      `json:"likes"`	Tags     []string `json:"tags,omitempty"`	Author   string   `json:"author,omitempty"`	BookName string   `json:"book_name,omitempty"`	BookID   string   `json:"book_id,omitempty"`	Text     string   `json:"text"`	ID       string   `json:"id"`
// //
// // type Quote struct {// Quote represents a quote from a book}	URL          string  `json:"url"`	LikesCount   int     `json:"likes_count"`	Text         string  `json:"text"`	Date         string  `json:"date,omitempty"`	Rating       int     `json:"rating"`	UserImageURL string  `json:"user_image_url,omitempty"`	UserName     string  `json:"user_name"`	BookID       string  `json:"book_id"`	ID           string  `json:"id"`
// //
// // type Review struct {// Review represents a book review}	RelatedAuthors  []string `json:"related_authors,omitempty"`	FansCount       int      `json:"fans_count,omitempty"`	ReviewCount     int      `json:"review_count,omitempty"`	RatingCount     int      `json:"rating_count,omitempty"`	AverageRating   float64  `json:"average_rating,omitempty"`	InfluencedBy    []string `json:"influenced_by,omitempty"`	Genres          []string `json:"genres,omitempty"`	Website         string   `json:"website,omitempty"`	DiedAt          string   `json:"died_at,omitempty"`	BornAt          string   `json:"born_at,omitempty"`	Bio             string   `json:"bio,omitempty"`	ImageURL        string   `json:"image_url,omitempty"`	URL             string   `json:"url,omitempty"`	Name            string   `json:"name"`	ID              string   `json:"id"`
// //
// // type Author struct {// Author represents an author from Goodreads}	URL            string   `json:"url"`
// // SeriesPosition string   `json:"series_position,omitempty"`
// // SeriesName     string   `json:"series_name,omitempty"`	Genres         []string `json:"genres,omitempty"`	CoverImageURL  string   `json:"cover_image_url,omitempty"`	Language       string   `json:"language,omitempty"`	Publisher      string   `json:"publisher,omitempty"`	PublishedYear  string   `json:"published_year,omitempty"`	PageCount      int      `json:"page_count,omitempty"`	ReviewCount    int      `json:"review_count"`	RatingCount    int      `json:"rating_count"`	Rating         float64  `json:"rating"`	Description    string   `json:"description,omitempty"`	ISBN13         string   `json:"isbn13,omitempty"`	ISBN           string   `json:"isbn,omitempty"`	Authors        []Author `json:"authors"`	Title          string   `json:"title"`	ID             string   `json:"id"`
