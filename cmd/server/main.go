package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/server"
	_ "modernc.org/sqlite"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		slog.Info("DATABASE_URL not set, using default path './db/bookscraping.db'")
		dbPath = "./db/bookscraping.db"
	}

	// Open database
	slog.Info("Opening database", "path", dbPath)
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
		slog.Info("SERVER_ADDR not set, using default address ':8080'")
	}

	// Start server
	srv := server.NewServer(
		server.WithQueries(queries),
		server.WithAddr(addr),
	)

	slog.Info("Starting BookScraping server",
		"address", addr,
		"database", dbPath,
	)

	if err := srv.Start(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
