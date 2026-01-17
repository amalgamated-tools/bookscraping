-- name: GetBook :one
SELECT * FROM books
WHERE id = ? LIMIT 1;

-- name: GetBookByBookID :one
SELECT * FROM books
WHERE book_id = ? LIMIT 1;

-- name: ListBooks :many
SELECT * FROM books
ORDER BY title ASC
LIMIT ? OFFSET ?;

-- name: CountBooks :one
SELECT COUNT(*) AS count FROM books;

-- name: CreateBook :one
INSERT INTO books (book_id, title, description, series_name, series_number, asin, isbn10, isbn13, language, hardcover_id, hardcover_book_id, goodreads_id, google_id, data, is_missing)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpsertBook :one
INSERT INTO books (book_id, title, description, series_name, series_number, asin, isbn10, isbn13, language, hardcover_id, hardcover_book_id, goodreads_id, google_id, data, is_missing)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
    data = excluded.data,
    is_missing = excluded.is_missing
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

-- name: UpsertAuthor :one
INSERT INTO authors (name)
VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=excluded.name
RETURNING *;

-- name: GetAuthorByName :one
SELECT * FROM authors
WHERE name = ? LIMIT 1;

-- name: LinkBookAuthor :exec
INSERT INTO book_authors (book_id, author_id)
VALUES (?, ?)
ON CONFLICT (book_id, author_id) DO NOTHING;

-- name: GetAuthorsForBook :many
SELECT a.id, a.name FROM authors a
JOIN book_authors ba ON a.id = ba.author_id
WHERE ba.book_id = ?
ORDER BY a.name ASC;

-- name: GetBooksBySeries :many
SELECT * FROM books
WHERE series_id = ?
ORDER BY series_number ASC;

-- name: GetSeriesAuthors :many
SELECT a.id, a.name FROM authors a
JOIN series_authors sa ON a.id = sa.author_id
WHERE sa.series_id = ?
ORDER BY a.name ASC;

-- name: LinkSeriesAuthor :exec
INSERT INTO series_authors (series_id, author_id)
VALUES (?, ?)
ON CONFLICT (series_id, author_id) DO NOTHING;

-- name: UpdateBookSeries :exec
UPDATE books
SET series_id = ?
WHERE id = ?;

-- name: CreateMissingBook :one
INSERT INTO books (book_id, title, description, series_name, series_number, goodreads_id, series_id, is_missing)
VALUES (?, ?, ?, ?, ?, ?, ?, 1)
ON CONFLICT(book_id) DO UPDATE SET
    title = excluded.title,
    description = excluded.description,
    series_name = excluded.series_name,
    series_number = excluded.series_number,
    goodreads_id = excluded.goodreads_id,
    series_id = excluded.series_id,
    is_missing = 1
RETURNING *;

-- name: GetSeriesByGoodreadsID :one
SELECT * FROM series
WHERE series_id = ? LIMIT 1;

-- name: GetConfig :one
SELECT value FROM configuration
WHERE key = ? LIMIT 1;

-- name: SetConfig :exec
INSERT INTO configuration (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;
