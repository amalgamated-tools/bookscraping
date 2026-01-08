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