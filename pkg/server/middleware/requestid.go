package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDLabel = "request_id"
const RequestID = "X-Request-ID"

// RequestIDKey is the context key for the X-Request-ID value
const ctxRequestIDKey = "go-http-RequestId"

// RequestIDHandler is a middleware that handles the  request id stuff
func RequestIDHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestID)

		if len(id) == 0 {
			if newID, err := uuid.NewRandom(); err == nil {
				id = newID.String()
			} else {
				id = "none"
			}
		}

		r = r.WithContext(WithRequestID(r.Context(), id))

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// GetRequestID returns the  request ID if one is present.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(ctxRequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// WithRequestID creates a new context based on the supplied parent, with the RequestId
// set to the specified value. If a RequestID already exists, it will be overwritten.
func WithRequestID(ctx context.Context, id string) context.Context {
	// nolint:staticcheck SA1029 should be ignored, context collisions are acceptable across major versions
	return context.WithValue(ctx, ctxRequestIDKey, id)
}

// Forward is a request hook that looks for X-Request-Id in the incoming context and adds them to r's headers.
func Forward(r *http.Request) {
	ctx := r.Context()

	if reqID := GetRequestID(ctx); reqID != "" {
		r.Header.Set(RequestID, reqID)
	}
}
