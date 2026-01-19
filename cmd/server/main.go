package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/server"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
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

	queries, err := db.SetupDatabase()
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
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
		server.WithQuerier(queries),
		server.WithAddr(addr),
	)

	slog.Info("Starting BookScraping server",
		slog.String("address", addr),
	)

	return srv.Run(cancelCtx)
}
