package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// UserRepository defines user persistence operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

// AccountRepository defines account persistence operations
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id AccountID) (*Account, error)
	GetByUserID(ctx context.Context, userID UserID) (*Account, error)
	Update(ctx context.Context, account *Account) error
	UpdateBalance(ctx context.Context, id AccountID, delta decimal.Decimal) error
}

// OrderRepository defines order persistence operations
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id OrderID) (*Order, error)
	GetByUserID(ctx context.Context, userID UserID, limit, offset int) ([]Order, error)
	GetPendingByUserID(ctx context.Context, userID UserID) ([]Order, error)
	GetPendingBySymbol(ctx context.Context, symbol string) ([]Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id OrderID) error
}

// PositionRepository defines position persistence operations
type PositionRepository interface {
	Create(ctx context.Context, position *Position) error
	GetByID(ctx context.Context, id PositionID) (*Position, error)
	GetByUserID(ctx context.Context, userID UserID) ([]Position, error)
	GetOpenByUserID(ctx context.Context, userID UserID) ([]Position, error)
	GetOpenByUserIDAndSymbol(ctx context.Context, userID UserID, symbol string) (*Position, error)
	GetAllOpen(ctx context.Context) ([]Position, error)
	GetOpenBySymbol(ctx context.Context, symbol string) ([]Position, error)
	Update(ctx context.Context, position *Position) error
	UpdatePnL(ctx context.Context, id PositionID, markPrice, unrealizedPnL decimal.Decimal) error
}

// TradeRepository defines trade persistence operations
type TradeRepository interface {
	Create(ctx context.Context, trade *Trade) error
	GetByID(ctx context.Context, id TradeID) (*Trade, error)
	GetByUserID(ctx context.Context, userID UserID, limit, offset int) ([]Trade, error)
	GetByPositionID(ctx context.Context, positionID PositionID) ([]Trade, error)
}

// PriceCache provides in-memory price lookups
type PriceCache interface {
	Get(symbol string) (*Price, bool)
	Set(symbol string, price *Price)
	GetAll() map[string]*Price
}
