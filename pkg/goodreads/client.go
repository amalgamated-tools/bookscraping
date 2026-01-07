package bookscraping

import "net/http"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Goodreads client
func NewClient() *Client {
	client := &http.Client{}
	client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	return &Client{
		baseURL:    "https://www.goodreads.com",
		httpClient: client,
	}
}
