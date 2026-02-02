package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderID int64

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

type Order struct {
	ID         OrderID
	UserID     UserID
	Symbol     string
	Side       OrderSide
	Type       OrderType
	Status     OrderStatus
	Quantity   decimal.Decimal
	Price      decimal.Decimal  // limit price (0 for market orders)
	Leverage   int
	StopLoss   *decimal.Decimal
	TakeProfit *decimal.Decimal
	FilledAt   *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsBuy returns true if this is a buy order
func (o *Order) IsBuy() bool {
	return o.Side == OrderSideBuy
}

// IsSell returns true if this is a sell order
func (o *Order) IsSell() bool {
	return o.Side == OrderSideSell
}

// IsMarket returns true if this is a market order
func (o *Order) IsMarket() bool {
	return o.Type == OrderTypeMarket
}

// IsLimit returns true if this is a limit order
func (o *Order) IsLimit() bool {
	return o.Type == OrderTypeLimit
}

// IsPending returns true if order is pending
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// CanBeCancelled returns true if order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending
}

// ToPositionSide converts order side to position side
func (o *Order) ToPositionSide() PositionSide {
	if o.Side == OrderSideBuy {
		return PositionSideLong
	}
	return PositionSideShort
}
