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
	err := c.ValidateToken(ctx)

	// err can be a generic error or one of ErrNoAccessToken/ErrInvalidToken
	switch err {
	case nil:
		methodLogger.Info("Access token valid, proceeding with sync")
	case ErrInvalidToken:
		methodLogger.Info("Access token invalid, attempting to refresh token")
		// if we have a refresh token, try to refresh the access token
		if c.accessToken.RefreshToken != "" {
			err = c.RefreshToken(ctx)
			if err != nil {
				methodLogger.Error("Failed to refresh access token during sync", slog.Any("error", err))
				// we can now try to login again
				err = c.performLogin(ctx)
				if err != nil {
					methodLogger.Error("Login failed during sync after refresh token failure", slog.Any("error", err))
					// return a wrapped error
					return fmt.Errorf("failed to login during sync after refresh token failure: %w", err)
				}
				methodLogger.Info("Login successful after refresh token failure, access token obtained")
			} else {
				methodLogger.Info("Access token refreshed successfully, proceeding with sync")
			}
		} else {
			methodLogger.Info("No refresh token available, attempting login to obtain access token")
			err = c.performLogin(ctx)
			if err != nil {
				methodLogger.Error("Login failed during sync", slog.Any("error", err))
				// return a wrapped error
				return fmt.Errorf("failed to login during sync: %w", err)
			}
			methodLogger.Info("Login successful, access token obtained")
		}
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
