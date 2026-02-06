package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PositionID int64

type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type PositionStatus string

const (
	PositionStatusOpen       PositionStatus = "OPEN"
	PositionStatusClosed     PositionStatus = "CLOSED"
	PositionStatusLiquidated PositionStatus = "LIQUIDATED"
)

type Position struct {
	ID               PositionID
	UserID           UserID
	Symbol           string
	Side             PositionSide
	Status           PositionStatus
	Quantity         decimal.Decimal // position size in base currency
	EntryPrice       decimal.Decimal // average entry price
	Leverage         int
	InitialMargin    decimal.Decimal // collateral locked
	MarkPrice        decimal.Decimal // current market price
	UnrealizedPnL    decimal.Decimal // current unrealized PnL
	RealizedPnL      decimal.Decimal // realized PnL (after close)
	LiquidationPrice decimal.Decimal
	StopLoss         *decimal.Decimal
	TakeProfit       *decimal.Decimal
	SLClosePercent   int // 1-100, default 100
	TPClosePercent   int // 1-100, default 100
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ClosedAt         *time.Time
}

// IsLong returns true if position is long
func (p *Position) IsLong() bool {
	return p.Side == PositionSideLong
}

// IsShort returns true if position is short
func (p *Position) IsShort() bool {
	return p.Side == PositionSideShort
}

// IsOpen returns true if position is open
func (p *Position) IsOpen() bool {
	return p.Status == PositionStatusOpen
}

// NotionalValue returns the notional value of the position
func (p *Position) NotionalValue() decimal.Decimal {
	return p.Quantity.Mul(p.MarkPrice)
}

// CalculatePnL calculates unrealized PnL at given price
// Long PnL:  Quantity * (MarkPrice - EntryPrice)
// Short PnL: Quantity * (EntryPrice - MarkPrice)
func (p *Position) CalculatePnL(markPrice decimal.Decimal) decimal.Decimal {
	if p.IsLong() {
		return p.Quantity.Mul(markPrice.Sub(p.EntryPrice))
	}
	return p.Quantity.Mul(p.EntryPrice.Sub(markPrice))
}

// ShouldLiquidate checks if position should be liquidated at given price
func (p *Position) ShouldLiquidate(markPrice decimal.Decimal) bool {
	if p.IsLong() {
		return markPrice.LessThanOrEqual(p.LiquidationPrice)
	}
	return markPrice.GreaterThanOrEqual(p.LiquidationPrice)
}

// ShouldTriggerStopLoss checks if stop loss should be triggered
func (p *Position) ShouldTriggerStopLoss(markPrice decimal.Decimal) bool {
	if p.StopLoss == nil {
		return false
	}
	if p.IsLong() {
		return markPrice.LessThanOrEqual(*p.StopLoss)
	}
	return markPrice.GreaterThanOrEqual(*p.StopLoss)
}

// ShouldTriggerTakeProfit checks if take profit should be triggered
func (p *Position) ShouldTriggerTakeProfit(markPrice decimal.Decimal) bool {
	if p.TakeProfit == nil {
		return false
	}
	if p.IsLong() {
		return markPrice.GreaterThanOrEqual(*p.TakeProfit)
	}
	return markPrice.LessThanOrEqual(*p.TakeProfit)
}

// UpdatePnL updates the position's mark price and unrealized PnL
func (p *Position) UpdatePnL(markPrice decimal.Decimal) {
	p.MarkPrice = markPrice
	p.UnrealizedPnL = p.CalculatePnL(markPrice)
}
