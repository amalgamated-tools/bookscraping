package booklore

import (
	"net/http"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
)

type Client struct {
	client *http.Client
	token  Token

	baseURL  string
	username string
	password string
}

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		client:   &http.Client{},
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// GetProjectRoot returns the root directory of the project.
func GetProjectRoot() string {
	return filepath.Join(filepath.Dir(b), "../..") //nolint:gocritic // This is a safe operation.
}
