package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/clevertrack1/mach"
)

type contextKey string

const requestIDKey contextKey = "requestID"

// generateRequestID generates a 128-bit random id.
func generateRequestID() string {
	var requestBytes [16]byte

	_, err := rand.Read(requestBytes[:])
	if err != nil {
		return ""
	}

	return hex.EncodeToString(requestBytes[:])
}

// RequestID adds a unique request ID to each request.
func RequestID() mach.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, httpReq *http.Request) {
			xRequestID := httpReq.Header.Get("X-Request-ID")

			if xRequestID == "" {
				xRequestID = generateRequestID()
			}

			// set context value and header
			ctx := context.WithValue(httpReq.Context(), requestIDKey, xRequestID)
			resp.Header().Set("X-Request-ID", xRequestID)

			next.ServeHTTP(resp, httpReq.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}

	return ""
}
