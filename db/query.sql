-- name: GetBook :one
SELECT * FROM books
WHERE id = ? LIMIT 1;

-- name: GetBookByBookID :one
SELECT * FROM books
WHERE book_id = ? LIMIT 1;

-- name: ListBooks :many
SELECT * FROM books
ORDER BY id ASC
LIMIT ? OFFSET ?;

-- name: CountBooks :one
SELECT COUNT(*) AS count FROM books;

-- name: CreateBook :one
INSERT INTO books (book_id, title, description, series_name, series_number, asin, isbn10, isbn13, language, hardcover_id, hardcover_book_id, goodreads_id, google_id, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpsertBook :one
INSERT INTO books (book_id, title, description, series_name, series_number, asin, isbn10, isbn13, language, hardcover_id, hardcover_book_id, goodreads_id, google_id, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(book_id) DO UPDATE SET
    title = excluded.title,
    description = excluded.description,
    series_name = excluded.series_name,
    series_number = excluded.series_number,
    asin = excluded.asin,
    isbn10 = excluded.isbn10,
    isbn13 = excluded.isbn13,
    language = excluded.language,
    hardcover_id = excluded.hardcover_id,
    hardcover_book_id = excluded.hardcover_book_id,
    goodreads_id = excluded.goodreads_id,
    google_id = excluded.google_id,
    data = excluded.data
RETURNING *;

-- name: GetSeries :one
SELECT * FROM series
WHERE id = ? LIMIT 1;

-- name: GetSeriesBySeriesID :one
SELECT * FROM series
WHERE series_id = ? LIMIT 1;

-- name: ListSeries :many
SELECT * FROM series
ORDER BY id ASC
LIMIT ? OFFSET ?;

-- name: CountSeries :one
SELECT COUNT(*) AS count FROM series;

-- name: CreateSeries :one
INSERT INTO series (series_id, name, description, url, data)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpsertSeries :one
INSERT INTO series (series_id, name, description, url, data)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(series_id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    url = excluded.url,
    data = excluded.data
RETURNING *;