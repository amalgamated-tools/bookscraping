package booklore

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

var (
	_, b, _, _ = runtime.Caller(0)
)

type Client struct {
	queries db.Querier
	client  *http.Client
	token   Token

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

func (c *Client) UpdateCredentials(baseURL, username, password string) {
	c.baseURL = baseURL
	c.username = username
	c.password = password
}

func (c *Client) GetToken() Token {
	return c.token
}

func (c *Client) SetToken(token Token) {
	c.token = token
}

// GetProjectRoot returns the root directory of the project.
func GetProjectRoot() string {
	return filepath.Join(filepath.Dir(b), "../..") //nolint:gocritic // This is a safe operation.
}
