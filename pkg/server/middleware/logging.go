package middleware

import (
	"log/slog"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(
			r.Context(),
			"Incoming request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("request_id", GetRequestID(r.Context())),
		)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
