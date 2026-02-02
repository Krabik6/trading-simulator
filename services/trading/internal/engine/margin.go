package engine

import (
	"github.com/shopspring/decimal"

	"trading/internal/domain"
)

// MarginCalculator handles margin-related calculations
type MarginCalculator struct {
	maintenanceRate decimal.Decimal
}

func NewMarginCalculator(maintenanceRate float64) *MarginCalculator {
	return &MarginCalculator{
		maintenanceRate: decimal.NewFromFloat(maintenanceRate),
	}
}

// CalculateInitialMargin calculates required margin for a position
// InitialMargin = (Quantity * Price) / Leverage
func (c *MarginCalculator) CalculateInitialMargin(quantity, price decimal.Decimal, leverage int) decimal.Decimal {
	notional := quantity.Mul(price)
	return notional.Div(decimal.NewFromInt(int64(leverage)))
}

// CalculateLiquidationPrice calculates the liquidation price for a position
// Long:  EntryPrice * (1 - 1/Leverage + MaintenanceRate)
// Short: EntryPrice * (1 + 1/Leverage - MaintenanceRate)
func (c *MarginCalculator) CalculateLiquidationPrice(entryPrice decimal.Decimal, leverage int, side domain.PositionSide) decimal.Decimal {
	leverageDec := decimal.NewFromInt(int64(leverage))
	leverageImpact := decimal.NewFromInt(1).Div(leverageDec)

	if side == domain.PositionSideLong {
		// Long: price drop triggers liquidation
		factor := decimal.NewFromInt(1).Sub(leverageImpact).Add(c.maintenanceRate)
		return entryPrice.Mul(factor)
	}

	// Short: price rise triggers liquidation
	factor := decimal.NewFromInt(1).Add(leverageImpact).Sub(c.maintenanceRate)
	return entryPrice.Mul(factor)
}

// CalculateRequiredMargin calculates total margin required including maintenance
func (c *MarginCalculator) CalculateRequiredMargin(quantity, price decimal.Decimal, leverage int) decimal.Decimal {
	initialMargin := c.CalculateInitialMargin(quantity, price, leverage)
	// Add small buffer for fees and maintenance
	buffer := initialMargin.Mul(c.maintenanceRate)
	return initialMargin.Add(buffer)
}

// HasSufficientMargin checks if account has enough margin for the position
func (c *MarginCalculator) HasSufficientMargin(
	availableMargin decimal.Decimal,
	quantity, price decimal.Decimal,
	leverage int,
) bool {
	required := c.CalculateRequiredMargin(quantity, price, leverage)
	return availableMargin.GreaterThanOrEqual(required)
}
