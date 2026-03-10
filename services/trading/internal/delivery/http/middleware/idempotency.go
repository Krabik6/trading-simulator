package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"trading/internal/domain"
	"trading/internal/idempotency"
	"trading/internal/logger"
)

type IdempotencyMiddleware struct {
	store idempotency.Store
}

func NewIdempotencyMiddleware(store idempotency.Store) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{store: store}
}

func (m *IdempotencyMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requiresIdempotency(r) {
			next.ServeHTTP(w, r)
			return
		}

		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key == "" {
			writeJSONError(w, "idempotency key is required", http.StatusBadRequest)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		requestHash := hashRequest(r.Method, r.URL.Path, bodyBytes)
		scope := r.Method + ":" + r.URL.Path
		userID := GetUserID(r.Context())

		acquireResult, err := m.store.Acquire(r.Context(), userID, scope, key, requestHash)
		if err != nil {
			switch {
			case err == domain.ErrIdempotencyKeyConflict:
				writeJSONError(w, "idempotency key already used with different request", http.StatusConflict)
			case err == domain.ErrIdempotencyInProgress:
				writeJSONError(w, "request with this idempotency key is already in progress", http.StatusConflict)
			default:
				logger.Error("idempotency acquire failed", "error", err)
				writeJSONError(w, "failed to process idempotency key", http.StatusInternalServerError)
			}
			return
		}

		if acquireResult.Replay != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotent-Replay", "true")
			w.WriteHeader(acquireResult.Replay.StatusCode)
			_, _ = w.Write([]byte(acquireResult.Replay.Body))
			return
		}

		recorder := &idempotencyResponseRecorder{ResponseWriter: w, status: http.StatusOK}
		completed := false
		defer func() {
			if recovered := recover(); recovered != nil {
				if !completed {
					if err := m.store.Complete(r.Context(), acquireResult.RecordID, http.StatusInternalServerError, `{"error":"internal server error"}`); err != nil {
						logger.Error("idempotency complete failed after panic", "error", err, "record_id", acquireResult.RecordID)
					}
				}
				panic(recovered)
			}
		}()

		next.ServeHTTP(recorder, r)

		if err := m.store.Complete(r.Context(), acquireResult.RecordID, recorder.status, recorder.body.String()); err != nil {
			logger.Error("idempotency complete failed", "error", err, "record_id", acquireResult.RecordID)
		}
		completed = true
	})
}

func requiresIdempotency(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
		return false
	}
	return strings.HasPrefix(r.URL.Path, "/orders") || strings.HasPrefix(r.URL.Path, "/positions")
}

func hashRequest(method, path string, body []byte) string {
	h := sha256.New()
	_, _ = h.Write([]byte(method))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(path))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

type idempotencyResponseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *idempotencyResponseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *idempotencyResponseRecorder) Write(p []byte) (int, error) {
	r.body.Write(p)
	return r.ResponseWriter.Write(p)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
