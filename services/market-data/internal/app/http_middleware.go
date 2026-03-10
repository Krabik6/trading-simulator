package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Krabik6/trading-simulator/market-data/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func httpMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		pathLabel := "unmatched"
		switch r.URL.Path {
		case "/health", "/ready":
			pathLabel = r.URL.Path
		}

		metrics.ObserveHTTPRequest(
			r.Method,
			pathLabel,
			strconv.Itoa(recorder.status),
			time.Since(start).Seconds(),
		)
	})
}
