package postgres

import (
	"context"
	"database/sql"
	"errors"

	"trading/internal/domain"
)

type OrderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (user_id, symbol, side, type, status, quantity, price, leverage, stop_loss, take_profit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		order.UserID, order.Symbol, order.Side, order.Type, order.Status,
		order.Quantity, order.Price, order.Leverage, order.StopLoss, order.TakeProfit,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

func (r *OrderRepository) GetByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, status, quantity, price, leverage,
			   stop_loss, take_profit, filled_at, created_at, updated_at
		FROM orders
		WHERE id = $1`

	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.Symbol, &order.Side, &order.Type,
		&order.Status, &order.Quantity, &order.Price, &order.Leverage,
		&order.StopLoss, &order.TakeProfit, &order.FilledAt,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID domain.UserID, limit, offset int) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, status, quantity, price, leverage,
			   stop_loss, take_profit, filled_at, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

func (r *OrderRepository) GetPendingByUserID(ctx context.Context, userID domain.UserID) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, status, quantity, price, leverage,
			   stop_loss, take_profit, filled_at, created_at, updated_at
		FROM orders
		WHERE user_id = $1 AND status = 'PENDING'
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

func (r *OrderRepository) GetPendingBySymbol(ctx context.Context, symbol string) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, status, quantity, price, leverage,
			   stop_loss, take_profit, filled_at, created_at, updated_at
		FROM orders
		WHERE symbol = $1 AND status = 'PENDING'
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1, filled_at = $2, quantity = $3, price = $4,
		    stop_loss = $5, take_profit = $6, updated_at = NOW()
		WHERE id = $7`

	result, err := r.db.ExecContext(ctx, query,
		order.Status, order.FilledAt, order.Quantity, order.Price,
		order.StopLoss, order.TakeProfit, order.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, id domain.OrderID) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) scanOrders(rows *sql.Rows) ([]domain.Order, error) {
	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Symbol, &order.Side, &order.Type,
			&order.Status, &order.Quantity, &order.Price, &order.Leverage,
			&order.StopLoss, &order.TakeProfit, &order.FilledAt,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}
