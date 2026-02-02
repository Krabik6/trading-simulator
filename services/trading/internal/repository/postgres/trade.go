package postgres

import (
	"context"
	"database/sql"
	"errors"

	"trading/internal/domain"
)

type TradeRepository struct {
	db *DB
}

func NewTradeRepository(db *DB) *TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) Create(ctx context.Context, trade *domain.Trade) error {
	query := `
		INSERT INTO trades (
			user_id, position_id, order_id, symbol, side, type,
			quantity, price, pnl, fee, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		trade.UserID, trade.PositionID, trade.OrderID, trade.Symbol,
		trade.Side, trade.Type, trade.Quantity, trade.Price,
		trade.PnL, trade.Fee,
	).Scan(&trade.ID, &trade.CreatedAt)
}

func (r *TradeRepository) GetByID(ctx context.Context, id domain.TradeID) (*domain.Trade, error) {
	query := `
		SELECT id, user_id, position_id, order_id, symbol, side, type,
			   quantity, price, pnl, fee, created_at
		FROM trades
		WHERE id = $1`

	trade := &domain.Trade{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trade.ID, &trade.UserID, &trade.PositionID, &trade.OrderID,
		&trade.Symbol, &trade.Side, &trade.Type,
		&trade.Quantity, &trade.Price, &trade.PnL, &trade.Fee,
		&trade.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTradeNotFound
		}
		return nil, err
	}
	return trade, nil
}

func (r *TradeRepository) GetByUserID(ctx context.Context, userID domain.UserID, limit, offset int) ([]domain.Trade, error) {
	query := `
		SELECT id, user_id, position_id, order_id, symbol, side, type,
			   quantity, price, pnl, fee, created_at
		FROM trades
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTrades(rows)
}

func (r *TradeRepository) GetByPositionID(ctx context.Context, positionID domain.PositionID) ([]domain.Trade, error) {
	query := `
		SELECT id, user_id, position_id, order_id, symbol, side, type,
			   quantity, price, pnl, fee, created_at
		FROM trades
		WHERE position_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, positionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTrades(rows)
}

func (r *TradeRepository) scanTrades(rows *sql.Rows) ([]domain.Trade, error) {
	var trades []domain.Trade
	for rows.Next() {
		var t domain.Trade
		err := rows.Scan(
			&t.ID, &t.UserID, &t.PositionID, &t.OrderID,
			&t.Symbol, &t.Side, &t.Type,
			&t.Quantity, &t.Price, &t.PnL, &t.Fee,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}
