package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trading/internal/domain"
	"trading/internal/idempotency"
)

const (
	idempotencyStatusProcessing = "PROCESSING"
	idempotencyStatusCompleted  = "COMPLETED"
	idempotencyStatusFailed     = "FAILED"
)

type IdempotencyStore struct {
	db *DB
}

func NewIdempotencyStore(db *DB) *IdempotencyStore {
	return &IdempotencyStore{db: db}
}

func (s *IdempotencyStore) Acquire(ctx context.Context, userID domain.UserID, scope, key, requestHash string) (*idempotency.AcquireResult, error) {
	var result *idempotency.AcquireResult

	err := s.db.WithinTx(ctx, func(txCtx context.Context) error {
		var recordID int64
		insertQuery := `
			INSERT INTO idempotency_keys (user_id, scope, key, request_hash, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (user_id, scope, key) DO NOTHING
			RETURNING id`

		err := s.db.QueryRowContext(txCtx, insertQuery, userID, scope, key, requestHash, idempotencyStatusProcessing).Scan(&recordID)
		if err == nil {
			result = &idempotency.AcquireResult{RecordID: recordID}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		var existingHash string
		var status string
		var statusCode sql.NullInt64
		var responseBody sql.NullString

		selectQuery := `
			SELECT id, request_hash, status, response_code, response_body
			FROM idempotency_keys
			WHERE user_id = $1 AND scope = $2 AND key = $3
			FOR UPDATE`

		err = s.db.QueryRowContext(txCtx, selectQuery, userID, scope, key).Scan(
			&recordID,
			&existingHash,
			&status,
			&statusCode,
			&responseBody,
		)
		if err != nil {
			return err
		}

		if existingHash != requestHash {
			return domain.ErrIdempotencyKeyConflict
		}

		switch status {
		case idempotencyStatusCompleted, idempotencyStatusFailed:
			code := 200
			if statusCode.Valid {
				code = int(statusCode.Int64)
			}
			body := "{}"
			if responseBody.Valid {
				body = responseBody.String
			}
			result = &idempotency.AcquireResult{
				RecordID: recordID,
				Replay: &idempotency.StoredResponse{
					StatusCode: code,
					Body:       body,
				},
			}
			return nil
		case idempotencyStatusProcessing:
			return domain.ErrIdempotencyInProgress
		default:
			return fmt.Errorf("unknown idempotency status %q", status)
		}
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *IdempotencyStore) Complete(ctx context.Context, recordID int64, statusCode int, body string) error {
	status := idempotencyStatusCompleted
	if statusCode >= 500 {
		status = idempotencyStatusFailed
	}

	query := `
		UPDATE idempotency_keys
		SET status = $1,
		    response_code = $2,
		    response_body = $3,
		    updated_at = NOW()
		WHERE id = $4`

	result, err := s.db.ExecContext(ctx, query, status, statusCode, body, recordID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrIdempotencyRecordNotFound
	}

	return nil
}
