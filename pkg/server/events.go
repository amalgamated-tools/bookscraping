package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for events
	clientEvents := make(chan string, 10)
	done := make(chan struct{})

	// Set up a ticker to send periodic updates (like a heartbeat)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Handle client disconnect
	go func() {
		<-r.Context().Done()
		close(done)
	}()

	// Send initial connection message
	fmt.Fprintf(w, "data: connected\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	slog.Info("SSE client connected")

	// Subscribe to server events
	go func() {
		for event := range s.eventCh {
			select {
			case <-done:
				return
			case clientEvents <- event:
			}
		}
	}()

	for {
		select {
		case <-done:
			slog.Info("SSE client disconnected")
			return
		case event := <-clientEvents:
			fmt.Fprintf(w, "data: %s\n\n", event)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			slog.Info("Sent SSE event to client", "event", event)
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

func (s *Server) handleTriggerEvent(w http.ResponseWriter, r *http.Request) {
	// Parse event message from request body
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Message == "" {
		writeError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Send the event to all connected SSE clients
	select {
	case s.eventCh <- payload.Message:
		slog.Info("Event triggered", "message", payload.Message)
		writeJSON(w, map[string]string{
			"status":  "success",
			"message": payload.Message,
		})
	case <-time.After(1 * time.Second):
		writeError(w, http.StatusInternalServerError, "Failed to send event")
	}
}
