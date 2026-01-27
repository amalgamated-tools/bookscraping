package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	_ "modernc.org/sqlite"
)

func main() {
	switch os.Args[1] {
	case "trigger":
		triggerCmd := flag.NewFlagSet("trigger", flag.ExitOnError)
		message := triggerCmd.String("message", "", "Message to send as SSE event")
		serverURL := triggerCmd.String("server", "http://localhost:8080", "Server URL")
		if err := triggerCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal("Failed to parse flags:", err)
		}

		if *message == "" {
			log.Fatal("--message flag is required")
		}
		triggerEvent(*serverURL, *message)
	case "migrate":
		_, err := db.SetupDatabase()
		if err != nil {
			log.Fatal("Failed to setup database:", err)
		}
		fmt.Println("Database migrated successfully")
	case "sync":
		queries, err := db.SetupDatabase()
		if err != nil {
			log.Fatal("Failed to setup database:", err)
		}
		client := booklore.NewClient(booklore.WithDBQueries(queries))
		client.Sync(context.Background())
	default:
		log.Fatal("Unknown command:", os.Args[1])
	}
}

func triggerEvent(serverURL, message string) {
	url := fmt.Sprintf("%s/api/events/trigger", serverURL)
	payload := map[string]string{"message": message}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("Failed to marshal payload:", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Failed to trigger event:", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal("Failed to close response body:", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body:", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to trigger event: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Println("Event triggered successfully:")
	fmt.Println(string(body))
}
