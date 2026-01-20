package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/stretchr/testify/mock"
)

func TestServer_handleGetConfig(t *testing.T) {
	tests := []struct {
		name             string
		getConfigFn      func(ctx context.Context, key string) (string, error)
		envVars          map[string]string
		expectedResponse ConfigRequest
	}{
		{
			name: "All values from database",
			getConfigFn: func(ctx context.Context, key string) (string, error) {
				switch key {
				case "serverUrl":
					return "https://booklore.example.com", nil
				case "username":
					return "testuser", nil
				case "password":
					return "testpass", nil
				}
				return "", nil
			},
			envVars: map[string]string{
				"BOOKLORE_SERVER":   "https://env.example.com",
				"BOOKLORE_USERNAME": "envuser",
				"BOOKLORE_PASSWORD": "envpass",
			},
			expectedResponse: ConfigRequest{
				ServerURL: "https://booklore.example.com",
				Username:  "testuser",
				Password:  "testpass",
			},
		},
		{
			name: "Fall back to environment variables",
			getConfigFn: func(ctx context.Context, key string) (string, error) {
				return "", nil // All database queries return empty
			},
			envVars: map[string]string{
				"BOOKLORE_SERVER":   "https://env.example.com",
				"BOOKLORE_USERNAME": "envuser",
				"BOOKLORE_PASSWORD": "envpass",
			},
			expectedResponse: ConfigRequest{
				ServerURL: "https://env.example.com",
				Username:  "envuser",
				Password:  "envpass",
			},
		},
		{
			name: "Partial database values with env fallback",
			getConfigFn: func(ctx context.Context, key string) (string, error) {
				switch key {
				case "serverUrl":
					return "https://booklore.example.com", nil
				case "username":
					return "", nil // Empty, should fall back to env
				case "password":
					return "dbpass", nil
				}
				return "", nil
			},
			envVars: map[string]string{
				"BOOKLORE_SERVER":   "https://env.example.com",
				"BOOKLORE_USERNAME": "envuser",
				"BOOKLORE_PASSWORD": "envpass",
			},
			expectedResponse: ConfigRequest{
				ServerURL: "https://booklore.example.com",
				Username:  "envuser",
				Password:  "dbpass",
			},
		},
		{
			name: "No database or environment values",
			getConfigFn: func(ctx context.Context, key string) (string, error) {
				return "", nil
			},
			envVars: map[string]string{},
			expectedResponse: ConfigRequest{
				ServerURL: "",
				Username:  "",
				Password:  "",
			},
		},
		{
			name: "Database error falls back to environment",
			getConfigFn: func(ctx context.Context, key string) (string, error) {
				return "", context.DeadlineExceeded // Simulate error
			},
			envVars: map[string]string{
				"BOOKLORE_SERVER":   "https://env.example.com",
				"BOOKLORE_USERNAME": "envuser",
				"BOOKLORE_PASSWORD": "envpass",
			},
			expectedResponse: ConfigRequest{
				ServerURL: "https://env.example.com",
				Username:  "envuser",
				Password:  "envpass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("Failed to unset environment variable %s: %v", key, err)
				}
			}
			defer func() {
				// Restore original environment
				for key := range tt.envVars {
					if err := os.Unsetenv(key); err != nil {
						t.Errorf("Failed to unset environment variable %s: %v", key, err)
					}
				}
				for key, val := range originalEnv {
					if val != "" {
						if err := os.Setenv(key, val); err != nil {
							t.Errorf("Failed to set environment variable %s: %v", key, err)
						}
					}
				}
			}()

			// Set test environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", key, err)
				}
			}
			mockQuerier := db.NewMockQuerier(t)
			mockQuerier.On("GetConfig", mock.Anything, mock.Anything).Return(tt.getConfigFn)
			_ = mockQuerier
			// Create server with mock queries
			server := &Server{
				queries: mockQuerier,
			}

			// Create request and response recorder
			req := httptest.NewRequest("GET", "/api/config", nil)
			w := httptest.NewRecorder()

			// Call handler
			server.handleGetConfig(w, req)

			// Check status code
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Check response body
			body, err := io.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var response ConfigRequest
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response != tt.expectedResponse {
				t.Errorf("Expected response %+v, got %+v", tt.expectedResponse, response)
			}

			// Check content-type header
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}
