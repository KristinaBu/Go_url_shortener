package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

type contextKey struct{}

func newRequestID() string {
	buffer := make([]byte, 16)
	_, _ = rand.Read(buffer)

	return hex.EncodeToString(buffer)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()

		ctx := context.WithValue(r.Context(), contextKey{}, requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(
			r.Context(),
			"http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		next.ServeHTTP(w, r)
	})
}
