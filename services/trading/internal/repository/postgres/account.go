package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (user_id, balance, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query, account.UserID, account.Balance).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *AccountRepository) GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID domain.UserID) (*domain.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1`

	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID, &account.UserID, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET balance = $1
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, account.Balance, account.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, id domain.AccountID, delta decimal.Decimal) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		RETURNING balance`

	var newBalance decimal.Decimal
	err := r.db.QueryRowContext(ctx, query, delta, id).Scan(&newBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return err
	}

	if newBalance.IsNegative() {
		// Rollback the change
		_, _ = r.db.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2`, delta, id)
		return domain.ErrInsufficientBalance
	}

	return nil
}
