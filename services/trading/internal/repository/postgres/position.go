package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

type PositionRepository struct {
	db *DB
}

func NewPositionRepository(db *DB) *PositionRepository {
	return &PositionRepository{db: db}
}

func (r *PositionRepository) Create(ctx context.Context, position *domain.Position) error {
	query := `
		INSERT INTO positions (
			user_id, symbol, side, status, quantity, entry_price, leverage,
			initial_margin, mark_price, unrealized_pnl, realized_pnl,
			liquidation_price, stop_loss, take_profit, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		position.UserID, position.Symbol, position.Side, position.Status,
		position.Quantity, position.EntryPrice, position.Leverage,
		position.InitialMargin, position.MarkPrice, position.UnrealizedPnL,
		position.RealizedPnL, position.LiquidationPrice,
		position.StopLoss, position.TakeProfit,
	).Scan(&position.ID, &position.CreatedAt, &position.UpdatedAt)
}

func (r *PositionRepository) GetByID(ctx context.Context, id domain.PositionID) (*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE id = $1`

	position := &domain.Position{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&position.ID, &position.UserID, &position.Symbol, &position.Side,
		&position.Status, &position.Quantity, &position.EntryPrice, &position.Leverage,
		&position.InitialMargin, &position.MarkPrice, &position.UnrealizedPnL,
		&position.RealizedPnL, &position.LiquidationPrice,
		&position.StopLoss, &position.TakeProfit,
		&position.CreatedAt, &position.UpdatedAt, &position.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPositionNotFound
		}
		return nil, err
	}
	return position, nil
}

func (r *PositionRepository) GetByUserID(ctx context.Context, userID domain.UserID) ([]domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

func (r *PositionRepository) GetOpenByUserID(ctx context.Context, userID domain.UserID) ([]domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE user_id = $1 AND status = 'OPEN'
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

func (r *PositionRepository) GetOpenByUserIDAndSymbol(ctx context.Context, userID domain.UserID, symbol string) (*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE user_id = $1 AND symbol = $2 AND status = 'OPEN'`

	position := &domain.Position{}
	err := r.db.QueryRowContext(ctx, query, userID, symbol).Scan(
		&position.ID, &position.UserID, &position.Symbol, &position.Side,
		&position.Status, &position.Quantity, &position.EntryPrice, &position.Leverage,
		&position.InitialMargin, &position.MarkPrice, &position.UnrealizedPnL,
		&position.RealizedPnL, &position.LiquidationPrice,
		&position.StopLoss, &position.TakeProfit,
		&position.CreatedAt, &position.UpdatedAt, &position.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPositionNotFound
		}
		return nil, err
	}
	return position, nil
}

func (r *PositionRepository) GetAllOpen(ctx context.Context) ([]domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE status = 'OPEN'`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

func (r *PositionRepository) GetOpenBySymbol(ctx context.Context, symbol string) ([]domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, side, status, quantity, entry_price, leverage,
			   initial_margin, mark_price, unrealized_pnl, realized_pnl,
			   liquidation_price, stop_loss, take_profit, created_at, updated_at, closed_at
		FROM positions
		WHERE symbol = $1 AND status = 'OPEN'`

	rows, err := r.db.QueryContext(ctx, query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

func (r *PositionRepository) Update(ctx context.Context, position *domain.Position) error {
	query := `
		UPDATE positions
		SET status = $1, quantity = $2, entry_price = $3, initial_margin = $4,
			mark_price = $5, unrealized_pnl = $6, realized_pnl = $7,
			liquidation_price = $8, stop_loss = $9, take_profit = $10, closed_at = $11
		WHERE id = $12`

	result, err := r.db.ExecContext(ctx, query,
		position.Status, position.Quantity, position.EntryPrice, position.InitialMargin,
		position.MarkPrice, position.UnrealizedPnL, position.RealizedPnL,
		position.LiquidationPrice, position.StopLoss, position.TakeProfit,
		position.ClosedAt, position.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrPositionNotFound
	}
	return nil
}

func (r *PositionRepository) UpdatePnL(ctx context.Context, id domain.PositionID, markPrice, unrealizedPnL decimal.Decimal) error {
	query := `
		UPDATE positions
		SET mark_price = $1, unrealized_pnl = $2
		WHERE id = $3 AND status = 'OPEN'`

	result, err := r.db.ExecContext(ctx, query, markPrice, unrealizedPnL, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrPositionNotFound
	}
	return nil
}

func (r *PositionRepository) scanPositions(rows *sql.Rows) ([]domain.Position, error) {
	var positions []domain.Position
	for rows.Next() {
		var p domain.Position
		err := rows.Scan(
			&p.ID, &p.UserID, &p.Symbol, &p.Side,
			&p.Status, &p.Quantity, &p.EntryPrice, &p.Leverage,
			&p.InitialMargin, &p.MarkPrice, &p.UnrealizedPnL,
			&p.RealizedPnL, &p.LiquidationPrice,
			&p.StopLoss, &p.TakeProfit,
			&p.CreatedAt, &p.UpdatedAt, &p.ClosedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}
