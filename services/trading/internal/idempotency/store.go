package idempotency

import (
	"context"

	"trading/internal/domain"
)

type StoredResponse struct {
	StatusCode int
	Body       string
}

type AcquireResult struct {
	RecordID int64
	Replay   *StoredResponse
}

type Store interface {
	Acquire(ctx context.Context, userID domain.UserID, scope, key, requestHash string) (*AcquireResult, error)
	Complete(ctx context.Context, recordID int64, statusCode int, body string) error
}
