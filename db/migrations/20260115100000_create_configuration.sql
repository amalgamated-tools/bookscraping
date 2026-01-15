-- migrate:up
CREATE TABLE configuration (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);

-- migrate:down
DROP TABLE configuration;
