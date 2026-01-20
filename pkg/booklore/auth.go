package booklore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	credentialsFile = filepath.Join(GetProjectRoot(), ".booklore_credentials.json")
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
func (c *Client) ValidateToken() error {
	if c.token.AccessToken == "" {
		return fmt.Errorf("no access token found")
	}
	url := c.baseURL + "/api/v1/users/me"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)

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
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (c *Client) RefreshToken() error {
	payload := strings.NewReader(`{"refreshToken": "` + c.token.RefreshToken + `"}`)
	url := c.baseURL + "/api/v1/auth/refresh"
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return err
	}

	c.token = token
	// save the token to credentials.json
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(credentialsFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) performLogin(ctx context.Context) error {
	methodLogger := slog.With(slog.String("method", "performLogin"))
	methodLogger.Debug("Performing login request to BookLore server...")

	payload := strings.NewReader(`{"username": "` + c.username + `", "password": "` + c.password + `"}`)
	url := c.baseURL + "/api/v1/auth/login"
	req, _ := http.NewRequestWithContext(ctx, "POST", url, payload)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// log the error type and message
		slog.Error("Login request failed", slog.Any("error", err), slog.String("error_type", fmt.Sprintf("%T", err)))
		return err
	}

	// log the response status
	slog.Info("Login response status", slog.String("status", res.Status))
	if res.StatusCode != 200 {
		slog.Error("Login failed", slog.String("status", res.Status))
		return fmt.Errorf("login failed with status: %s", res.Status)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read login response body", slog.Any("error", err))
		return err
	}

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return err
	}

	c.token = token

	return nil
}
