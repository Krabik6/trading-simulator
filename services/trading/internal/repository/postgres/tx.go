package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type txContextKey struct{}

func txFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx
}

func contextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx := txFromContext(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx := txFromContext(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx := txFromContext(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := txFromContext(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := contextWithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback transaction: %w (original error: %v)", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
