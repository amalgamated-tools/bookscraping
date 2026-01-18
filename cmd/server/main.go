package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/config"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/server"
	_ "modernc.org/sqlite"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Warn("Could not load config, using defaults", "error", err)
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./db/bookscraping.db"
	}

	// Open database
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("Failed to open database", "path", dbPath, "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	// Create queries instance
	queries := db.New(sqlDB)

	// Get server address
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Start server
	srv := server.NewServer(
		server.WithConfig(cfg),
		server.WithQueries(queries),
	)

	slog.Info("Starting BookScraping server",
		"address", addr,
		"database", dbPath,
	)

	_ = cfg // Use config if needed

	if err := srv.Start(addr); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
