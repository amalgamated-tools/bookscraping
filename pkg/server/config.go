package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

type ConfigRequest struct {
	ServerURL string `json:"serverUrl"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Try to get from database first
	serverUrl, err := s.queries.GetConfig(ctx, "serverUrl")
	if err != nil || serverUrl == "" {
		// Fall back to environment variable
		serverUrl = os.Getenv("BOOKLORE_SERVER")
	}

	username, err := s.queries.GetConfig(ctx, "username")
	if err != nil || username == "" {
		// Fall back to environment variable
		username = os.Getenv("BOOKLORE_USERNAME")
	}

	password, err := s.queries.GetConfig(ctx, "password")
	if err != nil || password == "" {
		// Fall back to environment variable
		password = os.Getenv("BOOKLORE_PASSWORD")
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

	ctx := r.Context()

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

func (s *Server) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ServerURL == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	ctx := r.Context()

	// Create a temporary client with the provided credentials
	client := booklore.NewClient(req.ServerURL, req.Username, req.Password)

	// Try to login
	if err := client.Login(); err != nil {
		slog.Error("Test connection failed", "error", err)
		writeError(w, http.StatusUnauthorized, "Connection failed: "+err.Error())
		return
	}

	// Get the token from the client
	token := client.GetToken()
	if token.AccessToken == "" {
		slog.Error("No access token returned from login")
		writeError(w, http.StatusInternalServerError, "No token returned from Booklore")
		return
	}

	// Store the access token in the database
	err := s.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   "booklore_access_token",
		Value: token.AccessToken,
	})
	if err != nil {
		slog.Error("Failed to store access token", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to store token")
		return
	}

	// Optionally store the refresh token as well
	if token.RefreshToken != "" {
		err := s.queries.SetConfig(ctx, db.SetConfigParams{
			Key:   "booklore_refresh_token",
			Value: token.RefreshToken,
		})
		if err != nil {
			slog.Warn("Failed to store refresh token", "error", err)
			// Don't fail the whole operation if refresh token storage fails
		}
	}

	writeJSON(w, map[string]string{"status": "success", "message": "Connection successful!"})
}
