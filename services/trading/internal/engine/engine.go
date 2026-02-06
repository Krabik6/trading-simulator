package engine

import (
	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

// Engine provides trading calculations
type Engine struct {
	MarginCalc      *MarginCalculator
	PnLCalc         *PnLCalculator
	LiquidationCalc *LiquidationChecker
	maxLeverage     int
}

func NewEngine(maxLeverage int, maintenanceRate float64) *Engine {
	marginCalc := NewMarginCalculator(maintenanceRate)
	return &Engine{
		MarginCalc:      marginCalc,
		PnLCalc:         NewPnLCalculator(),
		LiquidationCalc: NewLiquidationChecker(marginCalc),
		maxLeverage:     maxLeverage,
	}
}

// ValidateLeverage checks if leverage is within allowed range
func (e *Engine) ValidateLeverage(leverage int) bool {
	return leverage >= 1 && leverage <= e.maxLeverage
}

// ValidateStopLoss validates stop loss price for a position
func (e *Engine) ValidateStopLoss(stopLoss, entryPrice, liquidationPrice decimal.Decimal, side domain.PositionSide) error {
	if side == domain.PositionSideLong {
		// For long: SL must be below entry and above liquidation
		if stopLoss.GreaterThanOrEqual(entryPrice) {
			return domain.ErrInvalidStopLoss
		}
		if stopLoss.LessThanOrEqual(liquidationPrice) {
			return domain.ErrInvalidStopLoss
		}
	} else {
		// For short: SL must be above entry and below liquidation
		if stopLoss.LessThanOrEqual(entryPrice) {
			return domain.ErrInvalidStopLoss
		}
		if stopLoss.GreaterThanOrEqual(liquidationPrice) {
			return domain.ErrInvalidStopLoss
		}
	}
	return nil
}

// ValidateTakeProfit validates take profit price for a position
func (e *Engine) ValidateTakeProfit(takeProfit, entryPrice decimal.Decimal, side domain.PositionSide) error {
	if side == domain.PositionSideLong {
		// For long: TP must be above entry
		if takeProfit.LessThanOrEqual(entryPrice) {
			return domain.ErrInvalidTakeProfit
		}
	} else {
		// For short: TP must be below entry
		if takeProfit.GreaterThanOrEqual(entryPrice) {
			return domain.ErrInvalidTakeProfit
		}
	}
	return nil
}

// CreatePosition creates a new position with calculated values
func (e *Engine) CreatePosition(
	userID domain.UserID,
	symbol string,
	side domain.PositionSide,
	quantity, entryPrice decimal.Decimal,
	leverage int,
	stopLoss, takeProfit *decimal.Decimal,
) *domain.Position {
	initialMargin := e.MarginCalc.CalculateInitialMargin(quantity, entryPrice, leverage)
	liquidationPrice := e.MarginCalc.CalculateLiquidationPrice(entryPrice, leverage, side)

	return &domain.Position{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Status:           domain.PositionStatusOpen,
		Quantity:         quantity,
		EntryPrice:       entryPrice,
		Leverage:         leverage,
		InitialMargin:    initialMargin,
		MarkPrice:        entryPrice,
		UnrealizedPnL:    decimal.Zero,
		RealizedPnL:      decimal.Zero,
		LiquidationPrice: liquidationPrice,
		StopLoss:         stopLoss,
		TakeProfit:       takeProfit,
		SLClosePercent:   100,
		TPClosePercent:   100,
	}
}

// AddToPosition updates position when adding more quantity
func (e *Engine) AddToPosition(position *domain.Position, addQuantity, addPrice decimal.Decimal) {
	// Calculate new weighted average entry price
	newEntryPrice := e.PnLCalc.CalculateNewEntryPrice(
		position.Quantity, position.EntryPrice,
		addQuantity, addPrice,
	)

	// Update quantity
	newQuantity := position.Quantity.Add(addQuantity)

	// Recalculate margin and liquidation price
	newInitialMargin := e.MarginCalc.CalculateInitialMargin(newQuantity, newEntryPrice, position.Leverage)
	newLiquidationPrice := e.MarginCalc.CalculateLiquidationPrice(newEntryPrice, position.Leverage, position.Side)

	position.Quantity = newQuantity
	position.EntryPrice = newEntryPrice
	position.InitialMargin = newInitialMargin
	position.LiquidationPrice = newLiquidationPrice
}

// UpdatePositionPnL updates position's mark price and unrealized PnL
func (e *Engine) UpdatePositionPnL(position *domain.Position, markPrice decimal.Decimal) {
	position.MarkPrice = markPrice
	position.UnrealizedPnL = e.PnLCalc.CalculateUnrealizedPnL(position, markPrice)
}

// ClosePosition calculates realized PnL when closing a position
func (e *Engine) ClosePosition(position *domain.Position, closePrice decimal.Decimal) decimal.Decimal {
	return e.PnLCalc.CalculateRealizedPnL(position, closePrice)
}

// GetExecutionPrice returns the price at which an order should be executed
// Buy orders execute at ask price, sell orders at bid price
func (e *Engine) GetExecutionPrice(price *domain.Price, side domain.OrderSide) decimal.Decimal {
	if side == domain.OrderSideBuy {
		return decimal.NewFromFloat(price.Ask)
	}
	return decimal.NewFromFloat(price.Bid)
}
