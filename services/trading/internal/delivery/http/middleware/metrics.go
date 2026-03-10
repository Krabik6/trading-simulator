package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"trading/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// RequestMetrics captures per-request duration and status labels.
func RequestMetrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/metrics", "/health", "/ready":
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			routePattern := "unmatched"
			if rc := chi.RouteContext(r.Context()); rc != nil {
				if rp := rc.RoutePattern(); rp != "" {
					routePattern = rp
				}
			}

			metrics.ObserveHTTPRequest(
				r.Method,
				routePattern,
				strconv.Itoa(recorder.status),
				time.Since(start).Seconds(),
			)
		})
	}
}
