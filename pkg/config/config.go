package config

import "os"

type Config struct {
	BookloreServer   string
	BookloreUsername string
	BooklorePassword string
	DatabaseURL      string
}

func LoadConfig() (*Config, error) {
	c := &Config{}
	// Load Booklore config from environment variables (optional - can be set via web UI later)
	c.BookloreServer = os.Getenv("BOOKLORE_SERVER")
	c.BookloreUsername = os.Getenv("BOOKLORE_USERNAME")
	c.BooklorePassword = os.Getenv("BOOKLORE_PASSWORD")

	// Database URL is required
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	if c.DatabaseURL == "" {
		return nil, ErrMissingDatabaseURL
	}
	return c, nil
}

var (
	ErrMissingDatabaseURL = &ConfigError{"DATABASE_URL is required"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
