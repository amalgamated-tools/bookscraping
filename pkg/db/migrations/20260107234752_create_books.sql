-- migrate:up
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    series_name VARCHAR(255),
    series_number REAL,
    asin VARCHAR(13),
    isbn10 VARCHAR(10),
    isbn13 VARCHAR(13),
    language VARCHAR(10),
    hardcover_id VARCHAR(255),
    hardcover_book_id INT,
    goodreads_id VARCHAR(255),
    google_id VARCHAR(255),
    data JSON
)

-- migrate:down
DROP TABLE books;