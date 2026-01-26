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

	if err != nil {
		// token is invalid or expired, try to refresh it
		methodLogger.Info("Access token invalid or expired, refreshing token")
		err = c.RefreshToken()

		if err != nil {
			// we were unable to refresh the token, let's try to login again
			methodLogger.Info("Failed to refresh access token, performing login", slog.Any("error", err))
			err = c.performLogin(ctx)

			if err != nil {
				// the token was invalid, we couldn't refresh it, and login failed
				methodLogger.Error("Login failed during sync", slog.Any("error", err))
				// return a wrapped error
				return fmt.Errorf("failed to login during sync: %w", err)
			}

		}
		methodLogger.Info("Access token refreshed successfully")
	}
	return nil
}
