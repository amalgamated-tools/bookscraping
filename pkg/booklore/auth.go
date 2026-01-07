package booklore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	err := c.validateCredentials()
	if err == nil {
		// valid credentials found
		return nil
	}
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

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var token Token
	err := json.Unmarshal(body, &token)
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

func (c *Client) validateCredentials() error {
	// let's check if the credentials.json file exists
	info, err := os.Stat(credentialsFile)

	if err == nil && !info.IsDir() {
		// the file exists, read it
		data, err := os.ReadFile(credentialsFile)
		if err == nil {
			// the file was read successfully, unmarshal it
			var token Token
			err = json.Unmarshal(data, &token)
			if err == nil {
				c.token = token
				// validate the token
				err = c.ValidateToken()
				if err == nil {
					// the token is valid, return
					return nil
				}
				// the token is not valid, let's try to refresh it
				return c.RefreshToken()
			} // the file was not unmarshaled successfully
		} // the file was not read successfully
	} // the file does not exist or is invalid
	return fmt.Errorf("no valid credentials found")
}

func (c *Client) performLogin() error {
	payload := strings.NewReader(`{"username": "` + c.username + `", "password": "` + c.password + `"}`)
	url := c.baseURL + "/api/v1/auth/login"
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

	var token Token
	err := json.Unmarshal(body, &token)
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
