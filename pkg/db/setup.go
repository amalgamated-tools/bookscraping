package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
)

func SetupDatabase() (Querier, error) {
	slog.Info("Setting up database")
	
	// Determine database path: prefer mounted /data folder, fall back to ./db
	var dbFilePath string
	if _, err := os.Stat("/data"); err == nil {
		dbFilePath = "/data/bookscraping.db"
		slog.Info("Using mounted /data folder", slog.String("path", dbFilePath))
	} else {
		dbFilePath = "./db/bookscraping.db"
		slog.Info("Using local db folder", slog.String("path", dbFilePath))
	}
	
	// Ensure parent directory exists
	dbDir := filepath.Dir(dbFilePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", slog.String("path", dbDir), slog.Any("error", err))
		os.Exit(1)
	}
	
	// dbmate expects sqlite:path/to/db format
	dbmateURL := fmt.Sprintf("sqlite:%s", dbFilePath)
	parsedURL, err := url.Parse(dbmateURL)
	if err != nil {
		slog.Error("Failed to parse database URL", slog.String("url", dbmateURL), slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Parsed database URL", slog.String("scheme", parsedURL.Scheme), slog.String("path", parsedURL.Path))

	dbMate := dbmate.New(parsedURL)
	err = dbMate.CreateAndMigrate()
	if err != nil {
		slog.Error("Failed to create or migrate database", slog.String("path", dbFilePath), slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database created and migrated successfully", slog.String("path", dbFilePath))
	
	// Open database with sqlite driver (expects just the file path, not the URL format)
	slog.Info("Opening database", slog.String("path", dbFilePath))
	sqlDB, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		slog.Error("Failed to open database", slog.String("path", dbFilePath), slog.Any("error", err))
		os.Exit(1)
	}
	defer sqlDB.Close()

	// Create queries instance
	queries := New(sqlDB)
	count, err := queries.CountBooks(context.Background())
	if err != nil {
		slog.Error("Failed to count books in database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database connected", slog.Int64("book_count", count))
	return queries, nil
}
