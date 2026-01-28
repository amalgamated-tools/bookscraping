-- migrate:up
CREATE TABLE IF NOT EXISTS series (
    id INTEGER PRIMARY KEY,
    series_id INTEGER NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255),
    data JSON
)

-- migrate:down
DROP TABLE series;