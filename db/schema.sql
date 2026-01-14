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
);
CREATE TABLE series (
    id INTEGER PRIMARY KEY,
    series_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255),
    data JSON
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20260107234752'),
  ('20260107235102');
