package handler

import (
	"log/slog"
	"net/http"

	"github.com/KristinaBu/Go_url_shortener/pkg/requestid"
)

func LoggingMiddleware(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := requestid.New()

		ctx := requestid.WithID(r.Context(), requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		logger.InfoContext(
			ctx,
			"http request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		next.ServeHTTP(w, r)
	})
}
