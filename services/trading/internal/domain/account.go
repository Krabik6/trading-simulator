package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountID int64

type Account struct {
	ID        AccountID
	UserID    UserID
	Balance   decimal.Decimal // available balance (USDT)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountSummary contains calculated account metrics
type AccountSummary struct {
	Balance       decimal.Decimal // available balance
	Equity        decimal.Decimal // balance + unrealized PnL
	UsedMargin    decimal.Decimal // total margin used by open positions
	AvailableMargin decimal.Decimal // equity - used margin
	UnrealizedPnL decimal.Decimal // sum of all positions' unrealized PnL
	MarginRatio   decimal.Decimal // used margin / equity (for cross margin)
}

// CalculateSummary computes account summary with given positions
func (a *Account) CalculateSummary(positions []Position) AccountSummary {
	unrealizedPnL := decimal.Zero
	usedMargin := decimal.Zero

	for _, p := range positions {
		if p.Status == PositionStatusOpen {
			unrealizedPnL = unrealizedPnL.Add(p.UnrealizedPnL)
			usedMargin = usedMargin.Add(p.InitialMargin)
		}
	}

	equity := a.Balance.Add(unrealizedPnL)
	availableMargin := equity.Sub(usedMargin)

	var marginRatio decimal.Decimal
	if equity.IsPositive() {
		marginRatio = usedMargin.Div(equity)
	}

	return AccountSummary{
		Balance:         a.Balance,
		Equity:          equity,
		UsedMargin:      usedMargin,
		AvailableMargin: availableMargin,
		UnrealizedPnL:   unrealizedPnL,
		MarginRatio:     marginRatio,
	}
}
