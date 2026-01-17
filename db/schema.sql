CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE books (
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
, series_id INTEGER REFERENCES series(id) ON DELETE SET NULL, is_missing BOOLEAN DEFAULT 0);
CREATE TABLE series (
    id INTEGER PRIMARY KEY,
    series_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255),
    data JSON
);
CREATE TABLE authors (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);
CREATE TABLE book_authors (
    book_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    PRIMARY KEY (book_id, author_id),
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);
CREATE TABLE configuration (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);
CREATE INDEX idx_books_series_id ON books(series_id);
CREATE TABLE series_authors (
    series_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    PRIMARY KEY (series_id, author_id),
    FOREIGN KEY (series_id) REFERENCES series(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);
CREATE INDEX idx_series_authors_series_id ON series_authors(series_id);
CREATE INDEX idx_series_authors_author_id ON series_authors(author_id);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20260107234752'),
  ('20260107235102'),
  ('20260115002404'),
  ('20260115100000'),
  ('20260117154351'),
  ('20260117154352'),
  ('20260117155114');
