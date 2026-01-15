package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

type ConfigRequest struct {
	ServerURL string `json:"serverUrl"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	serverUrl, err := s.queries.GetConfig(ctx, "serverUrl")
	if err != nil {
		// If not found, just return empty strings, not an error
		serverUrl = ""
	}

	username, err := s.queries.GetConfig(ctx, "username")
	if err != nil {
		username = ""
	}

	password, err := s.queries.GetConfig(ctx, "password")
	if err != nil {
		password = ""
	}

	writeJSON(w, ConfigRequest{
		ServerURL: serverUrl,
		Username:  username,
		Password:  password,
	})
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()

	// Save each config item
	err := s.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   "serverUrl",
		Value: req.ServerURL,
	})
	if err != nil {
		slog.Error("Failed to save serverUrl", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	err = s.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   "username",
		Value: req.Username,
	})
	if err != nil {
		slog.Error("Failed to save username", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	err = s.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   "password",
		Value: req.Password,
	})
	if err != nil {
		slog.Error("Failed to save password", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	// Update the booklore client with new credentials
	s.blClient.UpdateCredentials(req.ServerURL, req.Username, req.Password)

	writeJSON(w, map[string]string{"status": "success"})
}
