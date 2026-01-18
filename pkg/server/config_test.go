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
)

// MockQuerier is a mock implementation of db.Querier for testing
type MockQuerier struct {
	getConfigFn func(ctx context.Context, key string) (string, error)
}

// GetConfig implements the GetConfig method from db.Querier
func (m *MockQuerier) GetConfig(ctx context.Context, key string) (string, error) {
	if m.getConfigFn != nil {
		return m.getConfigFn(ctx, key)
	}
	return "", nil
}

// Implement other required methods from Querier interface as no-ops
func (m *MockQuerier) CountBooks(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) CountSeries(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) CreateBook(ctx context.Context, arg db.CreateBookParams) (db.Book, error) {
	return db.Book{}, nil
}

func (m *MockQuerier) CreateMissingBook(ctx context.Context, arg db.CreateMissingBookParams) (db.Book, error) {
	return db.Book{}, nil
}

func (m *MockQuerier) CreateSeries(ctx context.Context, arg db.CreateSeriesParams) (db.Series, error) {
	return db.Series{}, nil
}

func (m *MockQuerier) GetAuthorByName(ctx context.Context, name string) (db.Author, error) {
	return db.Author{}, nil
}

func (m *MockQuerier) GetAuthorsForBook(ctx context.Context, bookID int64) ([]db.Author, error) {
	return nil, nil
}

func (m *MockQuerier) GetBook(ctx context.Context, id int64) (db.Book, error) {
	return db.Book{}, nil
}

func (m *MockQuerier) GetBookByBookID(ctx context.Context, bookID int64) (db.Book, error) {
	return db.Book{}, nil
}

func (m *MockQuerier) GetBooksBySeries(ctx context.Context, seriesID *int64) ([]db.Book, error) {
	return nil, nil
}

func (m *MockQuerier) GetSeries(ctx context.Context, id int64) (db.Series, error) {
	return db.Series{}, nil
}

func (m *MockQuerier) GetSeriesAuthors(ctx context.Context, seriesID int64) ([]db.Author, error) {
	return nil, nil
}

func (m *MockQuerier) GetSeriesByGoodreadsID(ctx context.Context, seriesID int64) (db.Series, error) {
	return db.Series{}, nil
}

func (m *MockQuerier) GetSeriesBySeriesID(ctx context.Context, seriesID int64) (db.Series, error) {
	return db.Series{}, nil
}

func (m *MockQuerier) LinkBookAuthor(ctx context.Context, arg db.LinkBookAuthorParams) error {
	return nil
}

func (m *MockQuerier) LinkSeriesAuthor(ctx context.Context, arg db.LinkSeriesAuthorParams) error {
	return nil
}

func (m *MockQuerier) ListBooks(ctx context.Context, arg db.ListBooksParams) ([]db.Book, error) {
	return nil, nil
}

func (m *MockQuerier) ListSeries(ctx context.Context, arg db.ListSeriesParams) ([]db.Series, error) {
	return nil, nil
}

func (m *MockQuerier) SetConfig(ctx context.Context, arg db.SetConfigParams) error {
	return nil
}

func (m *MockQuerier) UpdateBookSeries(ctx context.Context, arg db.UpdateBookSeriesParams) error {
	return nil
}

func (m *MockQuerier) UpsertAuthor(ctx context.Context, name string) (db.Author, error) {
	return db.Author{}, nil
}

func (m *MockQuerier) UpsertBook(ctx context.Context, arg db.UpsertBookParams) (db.Book, error) {
	return db.Book{}, nil
}

func (m *MockQuerier) UpsertSeries(ctx context.Context, arg db.UpsertSeriesParams) (db.Series, error) {
	return db.Series{}, nil
}

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
				os.Unsetenv(key)
			}
			defer func() {
				// Restore original environment
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
				for key, val := range originalEnv {
					if val != "" {
						os.Setenv(key, val)
					}
				}
			}()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create server with mock queries
			server := &Server{
				queries: &MockQuerier{
					getConfigFn: tt.getConfigFn,
				},
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
