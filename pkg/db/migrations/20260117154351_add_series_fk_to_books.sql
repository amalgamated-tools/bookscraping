-- migrate:up
ALTER TABLE books ADD COLUMN series_id INTEGER REFERENCES series(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_books_series_id ON books(series_id);

-- migrate:down
DROP INDEX IF EXISTS idx_books_series_id;
ALTER TABLE books DROP COLUMN series_id;
