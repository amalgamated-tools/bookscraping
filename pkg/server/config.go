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
	serverUrl, err := s.queries.GetConfig(ctx, db.ConfigKeyServerURL)
	if err != nil || serverUrl == "" {
		// Fall back to environment variable
		serverUrl = os.Getenv("BOOKLORE_SERVER")
	}

	username, err := s.queries.GetConfig(ctx, db.ConfigKeyUsername)
	if err != nil || username == "" {
		// Fall back to environment variable
		username = os.Getenv("BOOKLORE_USERNAME")
	}

	password, err := s.queries.GetConfig(ctx, db.ConfigKeyPassword)
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
	ctx := r.Context()
	methodLogger := slog.With(slog.String("handler", "handleSaveConfig"))

	methodLogger.DebugContext(ctx, "Unmarshaling request body")

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		methodLogger.ErrorContext(ctx, "Failed to decode request body", slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	methodLogger.DebugContext(ctx, "Unmarshaled request body")

	if req.ServerURL == "" || req.Username == "" || req.Password == "" {
		methodLogger.ErrorContext(ctx, "Missing credentials in request")
		writeError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	// Create a temporary client with the provided credentials
	client := booklore.NewClient(req.ServerURL, req.Username, req.Password)

	// Try to login
	if err := client.Login(ctx); err != nil {
		methodLogger.ErrorContext(ctx, "Test connection failed", slog.Any("error", err))
		writeError(w, http.StatusUnauthorized, "Connection failed: "+err.Error())
		return
	}

	// Get the token from the client
	token := client.GetToken()
	if token.AccessToken == "" {
		methodLogger.ErrorContext(ctx, "No access token returned from login")
		writeError(w, http.StatusInternalServerError, "No token returned from Booklore")
		return
	}

	// Store the configuration in the database
	for key, value := range map[string]string{
		db.ConfigKeyBookloreToken:    token.AccessToken,
		db.ConfigKeyBookloreRefToken: token.RefreshToken,
		db.ConfigKeyServerURL:        req.ServerURL,
		db.ConfigKeyUsername:         req.Username,
		db.ConfigKeyPassword:         req.Password,
	} {
		if err := s.queries.SetConfig(ctx, db.SetConfigParams{
			Key:   key,
			Value: value,
		}); err != nil {
			methodLogger.ErrorContext(ctx, "Failed to save config", slog.String("key", key), slog.Any("error", err))
			writeError(w, http.StatusInternalServerError, "Failed to save configuration")
			return
		}
	}

	// Update the booklore client with new credentials
	s.blClient = client

	writeJSON(w, map[string]string{"status": "success"})
}
