package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/server"
	"github.com/amalgamated-tools/bookscraping/pkg/telemetry"
)

var Version = "dev"

func main() {
	setupLogger()
	cancelCtx, cancelAll := context.WithCancel(context.Background())

	if err := realMain(cancelCtx); err != nil {
		fmt.Println(fmt.Errorf("\nerror: %w", err))
		cancelAll()
	}
}

func setupLogger() {
	format := "json"
	level := slog.LevelInfo

	logFormat, ok := os.LookupEnv("LOG_FORMAT")
	if ok {
		format = logFormat
	}

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		switch logLevel {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}
	}
	var logger *slog.Logger
	if format == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	// Try to get version from build info
	info, ok := debug.ReadBuildInfo()
	if ok {
		if info.Main.Version != "" {
			Version = info.Main.Version
		}
	}

	logger = logger.With(slog.String("version", Version))
	slog.SetDefault(logger)

	telemetry.Send(Version)
}

// This is the real main function. That's why it's called realMain.
func realMain(cancelCtx context.Context) error { //nolint:contextcheck // The newctx context comes from the StartTracer function, so it's already wrapped.
	flagSet := flag.NewFlagSet("http", flag.ExitOnError)

	var (
		port    int
		showVer bool
	)
	flagSet.IntVar(&port, "port", 0, "port number to run http server on")
	flagSet.BoolVar(&showVer, "version", false, "show version and exit")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if showVer {
		fmt.Println(Version)
		os.Exit(0)
	}

	queries, err := db.SetupDatabase()
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	// Get server address
	addr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		addr = ":8080"
		slog.Debug("SERVER_ADDR not set, using default address ':8080'")
	} else {
		slog.Debug("Using server address from SERVER_ADDR", slog.String("address", addr))
	}

	// Start server
	srv := server.NewServer(
		server.WithQuerier(queries),
		server.WithAddr(addr),
	)

	slog.Info("Starting BookScraping server",
		slog.String("address", addr),
	)

	return srv.Run(cancelCtx)
}
