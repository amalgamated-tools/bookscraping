package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

type ConfigRequest struct {
	ServerURL string `json:"serverUrl"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	methodLogger := slog.With(slog.String("method", "handleGetConfig"))
	ctx := r.Context()

	configs, err := db.GetAllConfig(ctx, s.queries)
	if err != nil {
		methodLogger.ErrorContext(ctx, "Failed to get all configs", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "Failed to get configuration")
		return
	}

	writeJSON(w, ConfigRequest{
		ServerURL: configs[db.BookloreServerURL],
		Username:  configs[db.BookloreUsername],
		Password:  configs[db.BooklorePassword],
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
	client := booklore.NewClient(
		ctx,
		booklore.WithCredentials(
			req.Username,
			req.Password,
		),
		booklore.WithBaseURL(req.ServerURL),
	)

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
		db.BookloreToken:     token.AccessToken,
		db.BookloreRefToken:  token.RefreshToken,
		db.BookloreServerURL: req.ServerURL,
		db.BookloreUsername:  req.Username,
		db.BooklorePassword:  req.Password,
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
