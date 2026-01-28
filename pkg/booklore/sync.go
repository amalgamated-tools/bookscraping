package booklore

import (
	"context"
	"fmt"
	"log/slog"
)

// Sync performs synchronization of book data with the server.
func (c *Client) Sync(ctx context.Context) error {
	methodLogger := slog.With(slog.String("method", "Sync"), slog.String("package", "booklore"))
	// we need to make sure that we have an access token that hasn't expired
	err := c.ValidateToken()

	// err can be a generic error or one of ErrNoAccessToken/ErrInvalidToken
	switch err {
	case nil:
		methodLogger.Info("Access token valid, proceeding with sync")
	case ErrInvalidToken:
		methodLogger.Info("Access token invalid, attempting to refresh token")
	case ErrNoAccessToken:
		methodLogger.Info("No access token found, attempting login to obtain access token")
		err = c.performLogin(ctx)

		if err != nil {
			methodLogger.Error("Login failed during sync", slog.Any("error", err))
			// return a wrapped error
			return fmt.Errorf("failed to login during sync: %w", err)
		}
		methodLogger.Info("Login successful, access token obtained")
	}

	return nil
}
