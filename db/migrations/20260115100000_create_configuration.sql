-- migrate:up
CREATE TABLE IF NOT EXISTS configuration (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);

-- migrate:down
DROP TABLE configuration;
