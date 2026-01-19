package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// Generate unique client ID
	clientID := uuid.New().String()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client's events
	clientEvents := make(chan string, 10)
	done := make(chan struct{})

	// Register client
	s.sseMu.Lock()
	s.sseClients[clientID] = clientEvents
	s.sseMu.Unlock()

	// Set up a ticker to send periodic updates (like a heartbeat)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Handle client disconnect
	go func() {
		<-r.Context().Done()
		close(done)
	}()

	// Send initial connection message with client ID
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clientId\":\"%s\"}\n\n", clientID)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	slog.Info("SSE client connected", slog.String("clientId", clientID))

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
			// Unregister client
			s.sseMu.Lock()
			delete(s.sseClients, clientID)
			s.sseMu.Unlock()
			slog.Info("SSE client disconnected", slog.String("clientId", clientID))
			return
		case event := <-clientEvents:
			fmt.Fprintf(w, "data: %s\n\n", event)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			slog.Info("Sent SSE event to client", slog.String("clientId", clientID), slog.String("event", event))
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
		Message  string `json:"message"`
		ClientID string `json:"clientId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Message == "" {
		writeError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Get active clients
	s.sseMu.RLock()
	clientCount := len(s.sseClients)

	// Send to specific client if specified
	if payload.ClientID != "" {
		if clientCh, exists := s.sseClients[payload.ClientID]; exists {
			s.sseMu.RUnlock()
			select {
			case clientCh <- payload.Message:
				slog.Info("Event triggered for specific client",
					slog.String("clientId", payload.ClientID),
					slog.String("message", payload.Message))
				writeJSON(w, map[string]any{
					"status":         "success",
					"message":        payload.Message,
					"clientId":       payload.ClientID,
					"recipientCount": 1,
				})
			case <-time.After(1 * time.Second):
				writeError(w, http.StatusInternalServerError, "Failed to send event to client")
			}
			return
		}
		s.sseMu.RUnlock()
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}

	// Send to all connected clients
	s.sseMu.RUnlock()

	// Send the event to all connected SSE clients
	select {
	case s.eventCh <- payload.Message:
		slog.Info("Event triggered for all clients",
			slog.String("message", payload.Message),
			slog.Int("recipientCount", clientCount))
		writeJSON(w, map[string]any{
			"status":         "success",
			"message":        payload.Message,
			"recipientCount": clientCount,
		})
	case <-time.After(1 * time.Second):
		writeError(w, http.StatusInternalServerError, "Failed to send event")
	}
}
