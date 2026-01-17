-- migrate:up
ALTER TABLE books ADD COLUMN is_missing BOOLEAN DEFAULT 0;

-- migrate:down
ALTER TABLE books DROP COLUMN is_missing;
