package booklore

import (
	"net/http"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

type ClientOption func(*Client)

func WithDBQueries(queries db.Querier) ClientOption {
	return func(c *Client) {
		c.queries = queries
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.client = httpClient
	}
}

func WithAccessToken(token Token) ClientOption {
	return func(c *Client) {
		c.accessToken = token
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}
