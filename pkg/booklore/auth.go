package booklore

import (
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
func (c *Client) Login() error {
	slog.Info("Logging in to BookLore...")
	return c.performLogin()
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

	defer res.Body.Close()
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

	defer res.Body.Close()
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

func (c *Client) performLogin() error {
	slog.Info("Performing login request to BookLore server...")
	payload := strings.NewReader(`{"username": "` + c.username + `", "password": "` + c.password + `"}`)
	url := c.baseURL + "/api/v1/auth/login"
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Login request failed", "error", err)
		return err
	}

	// log the response status
	slog.Info("Login response status", "status", res.Status)
	if res.StatusCode != 200 {
		slog.Error("Login failed", "status", res.Status)
		return fmt.Errorf("login failed with status: %s", res.Status)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read login response body", "error", err)
		return err
	}
	slog.Debug("Login response", "body", string(body))

	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		return err
	}

	c.token = token

	return nil
}
