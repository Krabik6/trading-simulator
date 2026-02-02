package engine

import (
	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

// PnLCalculator handles PnL calculations
type PnLCalculator struct{}

func NewPnLCalculator() *PnLCalculator {
	return &PnLCalculator{}
}

// CalculateUnrealizedPnL calculates unrealized PnL for a position
// Long PnL:  Quantity * (MarkPrice - EntryPrice)
// Short PnL: Quantity * (EntryPrice - MarkPrice)
func (c *PnLCalculator) CalculateUnrealizedPnL(position *domain.Position, markPrice decimal.Decimal) decimal.Decimal {
	if position.IsLong() {
		return position.Quantity.Mul(markPrice.Sub(position.EntryPrice))
	}
	return position.Quantity.Mul(position.EntryPrice.Sub(markPrice))
}

// CalculateRealizedPnL calculates realized PnL when closing a position
func (c *PnLCalculator) CalculateRealizedPnL(position *domain.Position, closePrice decimal.Decimal) decimal.Decimal {
	return c.CalculateUnrealizedPnL(position, closePrice)
}

// CalculateROE calculates Return on Equity (percentage)
// ROE = (PnL / InitialMargin) * 100
func (c *PnLCalculator) CalculateROE(pnl, initialMargin decimal.Decimal) decimal.Decimal {
	if initialMargin.IsZero() {
		return decimal.Zero
	}
	return pnl.Div(initialMargin).Mul(decimal.NewFromInt(100))
}

// CalculateNewEntryPrice calculates weighted average entry price when adding to position
// NewEntryPrice = (OldQuantity * OldPrice + NewQuantity * NewPrice) / (OldQuantity + NewQuantity)
func (c *PnLCalculator) CalculateNewEntryPrice(
	oldQuantity, oldPrice, newQuantity, newPrice decimal.Decimal,
) decimal.Decimal {
	totalQuantity := oldQuantity.Add(newQuantity)
	if totalQuantity.IsZero() {
		return decimal.Zero
	}

	totalValue := oldQuantity.Mul(oldPrice).Add(newQuantity.Mul(newPrice))
	return totalValue.Div(totalQuantity)
}
