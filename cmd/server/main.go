package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/server"
	_ "modernc.org/sqlite"
)

func main() {
	cancelCtx, cancelAll := context.WithCancel(context.Background())

	if err := realMain(cancelCtx); err != nil {
		fmt.Println(fmt.Errorf("\nerror: %w", err))
		// tools.FreakOut(cancelCtx, err, cancelAll)
		cancelAll()
	}
}

// This is the real main function. That's why it's called realMain.
func realMain(cancelCtx context.Context) error { //nolint:contextcheck // The newctx context comes from the StartTracer function, so it's already wrapped.
	flagSet := flag.NewFlagSet("http", flag.ExitOnError)

	var port int
	flagSet.IntVar(&port, "port", 0, "port number to run http server on")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Get database path from environment or use default
	dbPath, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		slog.Info("DATABASE_URL not set, using default path './db/bookscraping.db'")
		dbPath = "./db/bookscraping.db"
	} else {
		dbPath = strings.TrimPrefix(dbPath, "sqlite:")
		slog.Info("Using database path from DATABASE_URL", slog.String("path", dbPath))
	}

	// Open database
	slog.Info("Opening database", slog.String("path", dbPath))
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("Failed to open database", slog.String("path", dbPath), slog.Any("error", err))
		os.Exit(1)
	}
	defer sqlDB.Close()

	// Create queries instance
	queries := db.New(sqlDB)
	count, err := queries.CountBooks(context.Background())
	if err != nil {
		slog.Error("Failed to count books in database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database connected", slog.Int64("book_count", count))

	// Get server address
	addr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		addr = ":8080"
		slog.Info("SERVER_ADDR not set, using default address ':8080'")
	} else {
		slog.Info("Using server address from SERVER_ADDR", slog.String("address", addr))
	}

	// Start server
	srv := server.NewServer(
		server.WithQueries(queries),
		server.WithAddr(addr),
	)

	slog.Info("Starting BookScraping server",
		slog.String("address", addr),
		slog.String("database", dbPath),
	)

	return srv.Run(cancelCtx)
}
