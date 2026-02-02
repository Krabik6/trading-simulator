package engine

import (
	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

// LiquidationChecker handles liquidation checks
type LiquidationChecker struct {
	marginCalc *MarginCalculator
}

func NewLiquidationChecker(marginCalc *MarginCalculator) *LiquidationChecker {
	return &LiquidationChecker{
		marginCalc: marginCalc,
	}
}

// ShouldLiquidate checks if position should be liquidated at given mark price
func (c *LiquidationChecker) ShouldLiquidate(position *domain.Position, markPrice decimal.Decimal) bool {
	if position.IsLong() {
		// Long position: liquidate when price drops to or below liquidation price
		return markPrice.LessThanOrEqual(position.LiquidationPrice)
	}
	// Short position: liquidate when price rises to or above liquidation price
	return markPrice.GreaterThanOrEqual(position.LiquidationPrice)
}

// ShouldTriggerStopLoss checks if stop loss should be triggered
func (c *LiquidationChecker) ShouldTriggerStopLoss(position *domain.Position, markPrice decimal.Decimal) bool {
	if position.StopLoss == nil {
		return false
	}
	if position.IsLong() {
		return markPrice.LessThanOrEqual(*position.StopLoss)
	}
	return markPrice.GreaterThanOrEqual(*position.StopLoss)
}

// ShouldTriggerTakeProfit checks if take profit should be triggered
func (c *LiquidationChecker) ShouldTriggerTakeProfit(position *domain.Position, markPrice decimal.Decimal) bool {
	if position.TakeProfit == nil {
		return false
	}
	if position.IsLong() {
		return markPrice.GreaterThanOrEqual(*position.TakeProfit)
	}
	return markPrice.LessThanOrEqual(*position.TakeProfit)
}

// CheckPositionTriggers checks all triggers for a position
type TriggerResult struct {
	ShouldLiquidate   bool
	ShouldStopLoss    bool
	ShouldTakeProfit  bool
	TriggerPrice      decimal.Decimal
}

func (c *LiquidationChecker) CheckTriggers(position *domain.Position, markPrice decimal.Decimal) TriggerResult {
	result := TriggerResult{
		TriggerPrice: markPrice,
	}

	// Check liquidation first (highest priority)
	if c.ShouldLiquidate(position, markPrice) {
		result.ShouldLiquidate = true
		result.TriggerPrice = position.LiquidationPrice
		return result
	}

	// Check stop loss
	if c.ShouldTriggerStopLoss(position, markPrice) {
		result.ShouldStopLoss = true
		result.TriggerPrice = *position.StopLoss
		return result
	}

	// Check take profit
	if c.ShouldTriggerTakeProfit(position, markPrice) {
		result.ShouldTakeProfit = true
		result.TriggerPrice = *position.TakeProfit
		return result
	}

	return result
}
