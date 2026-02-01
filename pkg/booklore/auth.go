package booklore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

// Custom error types for Booklore authentication
var (
	ErrNoAccessToken      = errors.New("no access token found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrLoginFailed        = errors.New("login failed")
	ErrTokenRefreshFailed = errors.New("token refresh failed")
)

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Login performs login and saves the token to .booklore_credentials.json
// if the file exists, it reads the file and unmarshals into c.token and validates the token
func (c *Client) Login(ctx context.Context) error {
	slog.Info("Logging in to BookLore...")
	return c.performLogin(ctx)
}

// ValidateToken checks if the current token is valid by making a request to /api/v1/users/me
func (c *Client) ValidateToken(ctx context.Context) error {
	if c.accessToken.AccessToken == "" {
		return ErrNoAccessToken
	}
	url := c.baseURL + "/api/v1/users/me"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validate token request: %w", err)
	}

	req.Header.Add("accept", "*/*")
	req.Header.Add("Authorization", "Bearer "+c.accessToken.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()
	if res.StatusCode != 200 {
		return ErrInvalidToken
	}
	return nil
}

func (c *Client) RefreshToken(ctx context.Context) error {
	if c.accessToken.RefreshToken == "" {
		return ErrTokenRefreshFailed
	}
	payload := strings.NewReader(`{"refreshToken": "` + c.accessToken.RefreshToken + `"}`)
	url := c.baseURL + "/api/v1/auth/refresh"
	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return fmt.Errorf("failed to create refresh token request: %w", err)
	}

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh token request failed: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()

	if res.StatusCode != 200 {
		return ErrTokenRefreshFailed
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh token response: %w", err)
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return fmt.Errorf("failed to unmarshal refresh token: %w", err)
	}

	if token.AccessToken == "" || token.RefreshToken == "" {
		return ErrTokenRefreshFailed
	}

	c.accessToken = token

	if err := c.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   db.BookloreToken,
		Value: token.AccessToken,
	}); err != nil {
		// we will log this, but we can proceed
		slog.Error("Failed to save access token to db", slog.Any("error", err))
	}

	if err := c.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   db.BookloreRefToken,
		Value: token.RefreshToken,
	}); err != nil {
		// we will log this, but we can proceed
		slog.Error("Failed to save refresh token to db", slog.Any("error", err))
	}

	return nil
}

func (c *Client) performLogin(ctx context.Context) error {
	methodLogger := slog.With(slog.String("method", "performLogin"))

	if c.username == "" || c.password == "" || c.baseURL == "" {
		methodLogger.DebugContext(ctx, "Username, password, or baseURL not provided for login, loading from db")
		err := c.loadCredentials(ctx)
		if err != nil {
			methodLogger.ErrorContext(ctx, "Failed to load credentials from db", slog.Any("error", err))
			return fmt.Errorf("failed to load credentials from db: %w", err)
		}
	}

	if c.username == "" || c.password == "" || c.baseURL == "" {
		return fmt.Errorf("username, password, or baseURL not provided for login")
	}

	methodLogger.Debug("Performing login request to BookLore server...")

	payload := strings.NewReader(`{"username": "` + c.username + `", "password": "` + c.password + `"}`)
	url := c.baseURL + "/api/v1/auth/login"
	req, _ := http.NewRequestWithContext(ctx, "POST", url, payload)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Login request failed", slog.Any("error", err), slog.String("error_type", fmt.Sprintf("%T", err)))
		return fmt.Errorf("login request failed: %w", err)
	}

	slog.Info("Login response status", slog.String("status", res.Status))
	if res.StatusCode != 200 {
		slog.Error("Login failed", slog.String("status", res.Status))
		return fmt.Errorf("%w: %s", ErrLoginFailed, res.Status)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read login response body", slog.Any("error", err))
		return fmt.Errorf("failed to read login response body: %w", err)
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return fmt.Errorf("failed to unmarshal login token: %w", err)
	}

	if token.AccessToken == "" || token.RefreshToken == "" {
		return ErrLoginFailed
	}

	c.accessToken = token

	if err := c.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   db.BookloreToken,
		Value: token.AccessToken,
	}); err != nil {
		// we will log this, but we can proceed
		methodLogger.ErrorContext(ctx, "Failed to save access token to db", slog.Any("error", err))
	}

	if err := c.queries.SetConfig(ctx, db.SetConfigParams{
		Key:   db.BookloreRefToken,
		Value: token.RefreshToken,
	}); err != nil {
		// we will log this, but we can proceed
		methodLogger.ErrorContext(ctx, "Failed to save refresh token to db", slog.Any("error", err))
	}

	return nil
}

func (c *Client) loadCredentials(ctx context.Context) error {
	config, err := db.GetAllConfig(ctx, c.queries)
	if err != nil {
		return fmt.Errorf("failed to load credentials from db: %w", err)
	}

	c.username = config[db.BookloreUsername]
	c.password = config[db.BooklorePassword]
	c.baseURL = config[db.BookloreServerURL]
	c.accessToken = Token{
		AccessToken:  config[db.BookloreToken],
		RefreshToken: config[db.BookloreRefToken],
	}
	return nil
}
