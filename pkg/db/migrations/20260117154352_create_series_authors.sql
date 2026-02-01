-- migrate:up
CREATE TABLE IF NOT EXISTS series_authors (
    series_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    PRIMARY KEY (series_id, author_id),
    FOREIGN KEY (series_id) REFERENCES series(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_series_authors_series_id ON series_authors(series_id);
CREATE INDEX IF NOT EXISTS idx_series_authors_author_id ON series_authors(author_id);

-- migrate:down
DROP TABLE series_authors;
