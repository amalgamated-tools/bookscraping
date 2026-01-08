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
	c.BookloreServer = os.Getenv("BOOKLORE_SERVER")
	if c.BookloreServer == "" {
		return nil, ErrMissingBookloreServer
	}
	c.BookloreUsername = os.Getenv("BOOKLORE_USERNAME")
	if c.BookloreUsername == "" {
		return nil, ErrMissingBookloreUsername
	}
	c.BooklorePassword = os.Getenv("BOOKLORE_PASSWORD")
	if c.BooklorePassword == "" {
		return nil, ErrMissingBooklorePassword
	}
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	if c.DatabaseURL == "" {
		return nil, ErrMissingDatabaseURL
	}
	return c, nil
}

var (
	ErrMissingBookloreServer   = &ConfigError{"BOOKLORE_SERVER is required"}
	ErrMissingBookloreUsername = &ConfigError{"BOOKLORE_USERNAME is required"}
	ErrMissingBooklorePassword = &ConfigError{"BOOKLORE_PASSWORD is required"}
	ErrMissingDatabaseURL      = &ConfigError{"DATABASE_URL is required"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
