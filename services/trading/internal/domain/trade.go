package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type TradeID int64

type TradeType string

const (
	TradeTypeOpen      TradeType = "OPEN"
	TradeTypeClose     TradeType = "CLOSE"
	TradeTypeAdd       TradeType = "ADD"       // adding to existing position
	TradeTypeLiquidate TradeType = "LIQUIDATE"
)

type Trade struct {
	ID         TradeID
	UserID     UserID
	PositionID PositionID
	OrderID    OrderID
	Symbol     string
	Side       PositionSide
	Type       TradeType
	Quantity   decimal.Decimal
	Price      decimal.Decimal
	PnL        decimal.Decimal // realized PnL (for close/liquidate trades)
	Fee        decimal.Decimal
	CreatedAt  time.Time
}

// TradeEvent represents a trade event to be published to Kafka
type TradeEvent struct {
	TradeID    int64     `json:"trade_id"`
	UserID     int64     `json:"user_id"`
	PositionID int64     `json:"position_id"`
	OrderID    int64     `json:"order_id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Type       string    `json:"type"`
	Quantity   string    `json:"quantity"`
	Price      string    `json:"price"`
	PnL        string    `json:"pnl"`
	Fee        string    `json:"fee"`
	Timestamp  time.Time `json:"timestamp"`
}

// ToEvent converts Trade to TradeEvent for Kafka
func (t *Trade) ToEvent() TradeEvent {
	return TradeEvent{
		TradeID:    int64(t.ID),
		UserID:     int64(t.UserID),
		PositionID: int64(t.PositionID),
		OrderID:    int64(t.OrderID),
		Symbol:     t.Symbol,
		Side:       string(t.Side),
		Type:       string(t.Type),
		Quantity:   t.Quantity.String(),
		Price:      t.Price.String(),
		PnL:        t.PnL.String(),
		Fee:        t.Fee.String(),
		Timestamp:  t.CreatedAt,
	}
}
