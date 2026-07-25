package api

import (
	"net/http"
	"time"

	"github.com/2mes4/llull/internal/logging"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, traceID, spanID := logging.WithTraceContext(r.Context())
		r = r.WithContext(ctx)
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logging.Emit("info", r.Method+" "+r.URL.Path, traceID, spanID, &logging.Fields{
			Action:    &logging.Action{Type: "http", Name: r.Method},
			LatencyMs: time.Since(start).Milliseconds(),
		})
	})
}
