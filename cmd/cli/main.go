package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		// Default behavior - run the goodreads client test
		runGoodreadsTest()
		return
	}

	switch os.Args[1] {
	case "trigger":
		triggerCmd := flag.NewFlagSet("trigger", flag.ExitOnError)
		message := triggerCmd.String("message", "", "Message to send as SSE event")
		serverURL := triggerCmd.String("server", "http://localhost:8080", "Server URL")
		triggerCmd.Parse(os.Args[2:])

		if *message == "" {
			log.Fatal("--message flag is required")
		}
		triggerEvent(*serverURL, *message)
	case "goodreads":
		runGoodreadsTest()
	default:
		log.Fatal("Unknown command:", os.Args[1])
	}
}

func triggerEvent(serverURL, message string) {
	url := fmt.Sprintf("%s/api/events/trigger", serverURL)
	payload := map[string]string{"message": message}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Failed to trigger event:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to trigger event: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Println("Event triggered successfully:")
	fmt.Println(string(body))
}

func runGoodreadsTest() {
	grClient := goodreads.NewClient()
	series, err := grClient.GetSeries("40650")
	if err != nil {
		panic(err)
	}
	println("Series Title:", series.Title)
}
