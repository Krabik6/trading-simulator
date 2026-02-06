package position

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
	"trading/internal/engine"
	"trading/internal/logger"
	"trading/internal/metrics"
)

type UseCase struct {
	positionRepo domain.PositionRepository
	accountRepo  domain.AccountRepository
	tradeRepo    domain.TradeRepository
	orderRepo    domain.OrderRepository
	priceCache   domain.PriceCache
	engine       *engine.Engine
}

func NewUseCase(
	positionRepo domain.PositionRepository,
	accountRepo domain.AccountRepository,
	tradeRepo domain.TradeRepository,
	orderRepo domain.OrderRepository,
	priceCache domain.PriceCache,
	eng *engine.Engine,
) *UseCase {
	return &UseCase{
		positionRepo: positionRepo,
		accountRepo:  accountRepo,
		tradeRepo:    tradeRepo,
		orderRepo:    orderRepo,
		priceCache:   priceCache,
		engine:       eng,
	}
}

func (uc *UseCase) GetPositions(ctx context.Context, userID domain.UserID) ([]domain.Position, error) {
	return uc.positionRepo.GetOpenByUserID(ctx, userID)
}

func (uc *UseCase) GetPosition(ctx context.Context, userID domain.UserID, positionID domain.PositionID) (*domain.Position, error) {
	position, err := uc.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		return nil, err
	}

	if position.UserID != userID {
		return nil, domain.ErrPositionNotFound
	}

	return position, nil
}

type ClosePositionInput struct {
	UserID     domain.UserID
	PositionID domain.PositionID
	Quantity   *decimal.Decimal // nil = full close
}

func (uc *UseCase) ClosePosition(ctx context.Context, input ClosePositionInput) (*domain.Trade, error) {
	position, err := uc.positionRepo.GetByID(ctx, input.PositionID)
	if err != nil {
		return nil, err
	}

	if position.UserID != input.UserID {
		return nil, domain.ErrPositionNotFound
	}

	if !position.IsOpen() {
		return nil, domain.ErrPositionNotOpen
	}

	// Get current price
	price, ok := uc.priceCache.Get(position.Symbol)
	if !ok {
		return nil, domain.ErrPriceNotAvailable
	}

	// Determine close price (opposite side execution)
	var closePrice decimal.Decimal
	if position.IsLong() {
		closePrice = decimal.NewFromFloat(price.Bid) // sell at bid
	} else {
		closePrice = decimal.NewFromFloat(price.Ask) // buy at ask
	}

	// Partial close if quantity specified and less than position size
	if input.Quantity != nil && input.Quantity.IsPositive() && input.Quantity.LessThan(position.Quantity) {
		return uc.partialCloseAtPrice(ctx, position, closePrice, *input.Quantity, "user")
	}

	return uc.closePositionAtPrice(ctx, position, closePrice, "user")
}

func (uc *UseCase) closePositionAtPrice(
	ctx context.Context,
	position *domain.Position,
	closePrice decimal.Decimal,
	reason string,
) (*domain.Trade, error) {
	// Calculate realized PnL
	pnl := uc.engine.ClosePosition(position, closePrice)

	// Get account
	account, err := uc.accountRepo.GetByUserID(ctx, position.UserID)
	if err != nil {
		return nil, err
	}

	// Update position
	position.Status = domain.PositionStatusClosed
	position.RealizedPnL = pnl
	now := time.Now()
	position.ClosedAt = &now

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	// Credit only PnL to account (margin is virtual — never deducted on open)
	if err := uc.accountRepo.UpdateBalance(ctx, account.ID, pnl); err != nil {
		return nil, err
	}

	// Create a virtual order for the close
	closeSide := domain.OrderSideSell
	if position.IsShort() {
		closeSide = domain.OrderSideBuy
	}

	order := &domain.Order{
		UserID:   position.UserID,
		Symbol:   position.Symbol,
		Side:     closeSide,
		Type:     domain.OrderTypeMarket,
		Status:   domain.OrderStatusFilled,
		Quantity: position.Quantity,
		Price:    closePrice,
		Leverage: position.Leverage,
		FilledAt: &now,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Create trade record
	trade := &domain.Trade{
		UserID:     position.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     position.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeClose,
		Quantity:   position.Quantity,
		Price:      closePrice,
		PnL:        pnl,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	metrics.RecordPositionClosed(position.Symbol, string(position.Side), reason)

	logger.Info("position closed",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"side", position.Side,
		"pnl", pnl,
		"reason", reason,
	)

	return trade, nil
}

func (uc *UseCase) partialCloseAtPrice(
	ctx context.Context,
	position *domain.Position,
	closePrice decimal.Decimal,
	quantity decimal.Decimal,
	reason string,
) (*domain.Trade, error) {
	// Calculate proportional PnL
	proportion := quantity.Div(position.Quantity)
	fullPnL := uc.engine.ClosePosition(position, closePrice)
	pnl := fullPnL.Mul(proportion)

	// Get account
	account, err := uc.accountRepo.GetByUserID(ctx, position.UserID)
	if err != nil {
		return nil, err
	}

	// Reduce position
	marginRelease := position.InitialMargin.Mul(proportion)
	position.Quantity = position.Quantity.Sub(quantity)
	position.InitialMargin = position.InitialMargin.Sub(marginRelease)

	// Recalculate liquidation price (entry stays same)
	position.LiquidationPrice = uc.engine.MarginCalc.CalculateLiquidationPrice(
		position.EntryPrice, position.Leverage, position.Side,
	)

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	// Credit only PnL to account (margin is virtual)
	if err := uc.accountRepo.UpdateBalance(ctx, account.ID, pnl); err != nil {
		return nil, err
	}

	// Create a virtual order for the partial close
	closeSide := domain.OrderSideSell
	if position.IsShort() {
		closeSide = domain.OrderSideBuy
	}

	now := time.Now()
	order := &domain.Order{
		UserID:   position.UserID,
		Symbol:   position.Symbol,
		Side:     closeSide,
		Type:     domain.OrderTypeMarket,
		Status:   domain.OrderStatusFilled,
		Quantity: quantity,
		Price:    closePrice,
		Leverage: position.Leverage,
		FilledAt: &now,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Create trade record
	trade := &domain.Trade{
		UserID:     position.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     position.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeClose,
		Quantity:   quantity,
		Price:      closePrice,
		PnL:        pnl,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	logger.Info("position partially closed",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"side", position.Side,
		"closed_quantity", quantity,
		"remaining_quantity", position.Quantity,
		"pnl", pnl,
		"reason", reason,
	)

	return trade, nil
}

type UpdateTPSLInput struct {
	UserID         domain.UserID
	PositionID     domain.PositionID
	StopLoss       *decimal.Decimal
	TakeProfit     *decimal.Decimal
	SLClosePercent *int
	TPClosePercent *int
}

func (uc *UseCase) UpdateTPSL(ctx context.Context, input UpdateTPSLInput) (*domain.Position, error) {
	position, err := uc.positionRepo.GetByID(ctx, input.PositionID)
	if err != nil {
		return nil, err
	}

	if position.UserID != input.UserID {
		return nil, domain.ErrPositionNotFound
	}

	if !position.IsOpen() {
		return nil, domain.ErrPositionNotOpen
	}

	// Validate stop loss
	if input.StopLoss != nil {
		if err := uc.engine.ValidateStopLoss(*input.StopLoss, position.EntryPrice, position.LiquidationPrice, position.Side); err != nil {
			return nil, err
		}
		position.StopLoss = input.StopLoss
	}

	// Validate take profit
	if input.TakeProfit != nil {
		if err := uc.engine.ValidateTakeProfit(*input.TakeProfit, position.EntryPrice, position.Side); err != nil {
			return nil, err
		}
		position.TakeProfit = input.TakeProfit
	}

	// Update close percents
	if input.SLClosePercent != nil {
		v := *input.SLClosePercent
		if v < 1 || v > 100 {
			return nil, domain.ErrInvalidClosePercent
		}
		position.SLClosePercent = v
	}
	if input.TPClosePercent != nil {
		v := *input.TPClosePercent
		if v < 1 || v > 100 {
			return nil, domain.ErrInvalidClosePercent
		}
		position.TPClosePercent = v
	}

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	logger.Info("position TP/SL updated",
		"position_id", position.ID,
		"stop_loss", position.StopLoss,
		"take_profit", position.TakeProfit,
	)

	return position, nil
}

// Liquidate liquidates a position (called by price processor)
func (uc *UseCase) Liquidate(ctx context.Context, position *domain.Position, liquidationPrice decimal.Decimal) (*domain.Trade, error) {
	if !position.IsOpen() {
		return nil, domain.ErrPositionNotOpen
	}

	account, err := uc.accountRepo.GetByUserID(ctx, position.UserID)
	if err != nil {
		return nil, err
	}

	// At liquidation, the user loses their initial margin
	// PnL = -InitialMargin (simplified)
	pnl := position.InitialMargin.Neg()

	// Update position
	position.Status = domain.PositionStatusLiquidated
	position.RealizedPnL = pnl
	now := time.Now()
	position.ClosedAt = &now

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	// Margin was virtual (never deducted from balance on open).
	// On liquidation the user loses the full margin, so we deduct it now.

	// Create order for record
	closeSide := domain.OrderSideSell
	if position.IsShort() {
		closeSide = domain.OrderSideBuy
	}

	order := &domain.Order{
		UserID:   position.UserID,
		Symbol:   position.Symbol,
		Side:     closeSide,
		Type:     domain.OrderTypeMarket,
		Status:   domain.OrderStatusFilled,
		Quantity: position.Quantity,
		Price:    liquidationPrice,
		Leverage: position.Leverage,
		FilledAt: &now,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	trade := &domain.Trade{
		UserID:     position.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     position.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeLiquidate,
		Quantity:   position.Quantity,
		Price:      liquidationPrice,
		PnL:        pnl,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	metrics.RecordLiquidation(position.Symbol)
	metrics.RecordPositionClosed(position.Symbol, string(position.Side), "liquidation")

	// Deduct margin from account
	if err := uc.accountRepo.UpdateBalance(ctx, account.ID, pnl); err != nil {
		logger.Error("failed to deduct liquidation loss", "error", err)
	}

	logger.Warn("position liquidated",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"side", position.Side,
		"liquidation_price", liquidationPrice,
		"loss", pnl,
	)

	return trade, nil
}

// TriggerStopLoss closes position at stop loss price (fully or partially based on SLClosePercent)
func (uc *UseCase) TriggerStopLoss(ctx context.Context, position *domain.Position) (*domain.Trade, error) {
	if position.StopLoss == nil {
		return nil, nil
	}

	pct := position.SLClosePercent
	if pct <= 0 {
		pct = 100
	}

	if pct < 100 {
		closeQty := position.Quantity.Mul(decimal.NewFromInt(int64(pct))).Div(decimal.NewFromInt(100))
		trade, err := uc.partialCloseAtPrice(ctx, position, *position.StopLoss, closeQty, "stop_loss")
		if err != nil {
			return nil, err
		}
		// Clear SL after partial trigger to prevent repeated firing
		position.StopLoss = nil
		position.SLClosePercent = 100
		if err := uc.positionRepo.Update(ctx, position); err != nil {
			return nil, err
		}
		return trade, nil
	}

	return uc.closePositionAtPrice(ctx, position, *position.StopLoss, "stop_loss")
}

// TriggerTakeProfit closes position at take profit price (fully or partially based on TPClosePercent)
func (uc *UseCase) TriggerTakeProfit(ctx context.Context, position *domain.Position) (*domain.Trade, error) {
	if position.TakeProfit == nil {
		return nil, nil
	}

	pct := position.TPClosePercent
	if pct <= 0 {
		pct = 100
	}

	if pct < 100 {
		closeQty := position.Quantity.Mul(decimal.NewFromInt(int64(pct))).Div(decimal.NewFromInt(100))
		trade, err := uc.partialCloseAtPrice(ctx, position, *position.TakeProfit, closeQty, "take_profit")
		if err != nil {
			return nil, err
		}
		// Clear TP after partial trigger to prevent repeated firing
		position.TakeProfit = nil
		position.TPClosePercent = 100
		if err := uc.positionRepo.Update(ctx, position); err != nil {
			return nil, err
		}
		return trade, nil
	}

	return uc.closePositionAtPrice(ctx, position, *position.TakeProfit, "take_profit")
}
