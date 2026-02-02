package order

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
	"trading/internal/engine"
	"trading/internal/logger"
	"trading/internal/metrics"
)

type PlaceOrderInput struct {
	UserID     domain.UserID
	Symbol     string
	Side       domain.OrderSide
	Type       domain.OrderType
	Quantity   decimal.Decimal
	Price      decimal.Decimal // for limit orders
	Leverage   int
	StopLoss   *decimal.Decimal
	TakeProfit *decimal.Decimal
}

type PlaceOrderOutput struct {
	Order    *domain.Order
	Position *domain.Position
	Trade    *domain.Trade
}

type UseCase struct {
	orderRepo    domain.OrderRepository
	positionRepo domain.PositionRepository
	accountRepo  domain.AccountRepository
	tradeRepo    domain.TradeRepository
	priceCache   domain.PriceCache
	engine       *engine.Engine
	symbols      map[string]bool
}

func NewUseCase(
	orderRepo domain.OrderRepository,
	positionRepo domain.PositionRepository,
	accountRepo domain.AccountRepository,
	tradeRepo domain.TradeRepository,
	priceCache domain.PriceCache,
	eng *engine.Engine,
	supportedSymbols []string,
) *UseCase {
	symbols := make(map[string]bool)
	for _, s := range supportedSymbols {
		symbols[s] = true
	}
	return &UseCase{
		orderRepo:    orderRepo,
		positionRepo: positionRepo,
		accountRepo:  accountRepo,
		tradeRepo:    tradeRepo,
		priceCache:   priceCache,
		engine:       eng,
		symbols:      symbols,
	}
}

func (uc *UseCase) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*PlaceOrderOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Get current price
	price, ok := uc.priceCache.Get(input.Symbol)
	if !ok {
		return nil, domain.ErrPriceNotAvailable
	}

	// Get account
	account, err := uc.accountRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Get existing position for this symbol
	existingPosition, err := uc.positionRepo.GetOpenByUserIDAndSymbol(ctx, input.UserID, input.Symbol)
	if err != nil && !errors.Is(err, domain.ErrPositionNotFound) {
		return nil, err
	}

	// Determine execution price
	executionPrice := uc.engine.GetExecutionPrice(price, input.Side)

	// For limit orders, use specified price
	if input.Type == domain.OrderTypeLimit {
		executionPrice = input.Price
	}

	// Calculate required margin
	requiredMargin := uc.engine.MarginCalc.CalculateRequiredMargin(input.Quantity, executionPrice, input.Leverage)

	// Get open positions for margin calculation
	openPositions, err := uc.positionRepo.GetOpenByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Calculate available margin
	summary := account.CalculateSummary(openPositions)

	// Check if we have enough margin
	if summary.AvailableMargin.LessThan(requiredMargin) {
		return nil, domain.ErrInsufficientMargin
	}

	// Create order
	order := &domain.Order{
		UserID:     input.UserID,
		Symbol:     input.Symbol,
		Side:       input.Side,
		Type:       input.Type,
		Status:     domain.OrderStatusPending,
		Quantity:   input.Quantity,
		Price:      executionPrice,
		Leverage:   input.Leverage,
		StopLoss:   input.StopLoss,
		TakeProfit: input.TakeProfit,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	metrics.RecordOrderPlaced(input.Symbol, string(input.Side), string(input.Type))

	// For market orders, execute immediately
	if input.Type == domain.OrderTypeMarket {
		return uc.executeOrder(ctx, order, existingPosition, executionPrice, account)
	}

	// Limit order stays pending
	return &PlaceOrderOutput{Order: order}, nil
}

func (uc *UseCase) executeOrder(
	ctx context.Context,
	order *domain.Order,
	existingPosition *domain.Position,
	executionPrice decimal.Decimal,
	account *domain.Account,
) (*PlaceOrderOutput, error) {
	now := time.Now()
	order.Status = domain.OrderStatusFilled
	order.FilledAt = &now

	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	positionSide := order.ToPositionSide()

	// Check if we need to close/reduce existing position or open/add to one
	if existingPosition != nil {
		if existingPosition.Side == positionSide {
			// Same direction - add to position
			return uc.addToPosition(ctx, order, existingPosition, executionPrice)
		}
		// Opposite direction - close or reduce position
		return uc.reducePosition(ctx, order, existingPosition, executionPrice, account)
	}

	// No existing position - open new one
	return uc.openPosition(ctx, order, executionPrice)
}

func (uc *UseCase) openPosition(
	ctx context.Context,
	order *domain.Order,
	executionPrice decimal.Decimal,
) (*PlaceOrderOutput, error) {
	position := uc.engine.CreatePosition(
		order.UserID,
		order.Symbol,
		order.ToPositionSide(),
		order.Quantity,
		executionPrice,
		order.Leverage,
		order.StopLoss,
		order.TakeProfit,
	)

	if err := uc.positionRepo.Create(ctx, position); err != nil {
		return nil, err
	}

	trade := &domain.Trade{
		UserID:     order.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     order.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeOpen,
		Quantity:   order.Quantity,
		Price:      executionPrice,
		PnL:        decimal.Zero,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	metrics.RecordOrderFilled(order.Symbol, string(order.Side))
	metrics.RecordPositionOpened(order.Symbol, string(position.Side))

	logger.Info("position opened",
		"position_id", position.ID,
		"symbol", order.Symbol,
		"side", position.Side,
		"quantity", order.Quantity,
		"entry_price", executionPrice,
		"leverage", order.Leverage,
	)

	return &PlaceOrderOutput{
		Order:    order,
		Position: position,
		Trade:    trade,
	}, nil
}

func (uc *UseCase) addToPosition(
	ctx context.Context,
	order *domain.Order,
	position *domain.Position,
	executionPrice decimal.Decimal,
) (*PlaceOrderOutput, error) {
	// Add to existing position
	uc.engine.AddToPosition(position, order.Quantity, executionPrice)

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	trade := &domain.Trade{
		UserID:     order.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     order.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeAdd,
		Quantity:   order.Quantity,
		Price:      executionPrice,
		PnL:        decimal.Zero,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	metrics.RecordOrderFilled(order.Symbol, string(order.Side))

	logger.Info("added to position",
		"position_id", position.ID,
		"added_quantity", order.Quantity,
		"new_entry_price", position.EntryPrice,
		"new_quantity", position.Quantity,
	)

	return &PlaceOrderOutput{
		Order:    order,
		Position: position,
		Trade:    trade,
	}, nil
}

func (uc *UseCase) reducePosition(
	ctx context.Context,
	order *domain.Order,
	position *domain.Position,
	executionPrice decimal.Decimal,
	account *domain.Account,
) (*PlaceOrderOutput, error) {
	// Calculate realized PnL for the closed portion
	closeQuantity := order.Quantity
	if closeQuantity.GreaterThan(position.Quantity) {
		closeQuantity = position.Quantity
	}

	// Calculate proportional PnL
	proportion := closeQuantity.Div(position.Quantity)
	pnl := uc.engine.PnLCalc.CalculateUnrealizedPnL(position, executionPrice).Mul(proportion)

	// Calculate margin to release
	marginRelease := position.InitialMargin.Mul(proportion)

	// Update position or close it
	if closeQuantity.Equal(position.Quantity) {
		// Full close
		position.Status = domain.PositionStatusClosed
		position.RealizedPnL = pnl
		now := time.Now()
		position.ClosedAt = &now
	} else {
		// Partial close - reduce position
		position.Quantity = position.Quantity.Sub(closeQuantity)
		position.InitialMargin = position.InitialMargin.Sub(marginRelease)
		// Recalculate liquidation price (entry stays same)
		position.LiquidationPrice = uc.engine.MarginCalc.CalculateLiquidationPrice(
			position.EntryPrice, position.Leverage, position.Side,
		)
	}

	if err := uc.positionRepo.Update(ctx, position); err != nil {
		return nil, err
	}

	// Credit PnL + margin to account
	totalCredit := pnl.Add(marginRelease)
	if err := uc.accountRepo.UpdateBalance(ctx, account.ID, totalCredit); err != nil {
		return nil, err
	}

	trade := &domain.Trade{
		UserID:     order.UserID,
		PositionID: position.ID,
		OrderID:    order.ID,
		Symbol:     order.Symbol,
		Side:       position.Side,
		Type:       domain.TradeTypeClose,
		Quantity:   closeQuantity,
		Price:      executionPrice,
		PnL:        pnl,
		Fee:        decimal.Zero,
	}

	if err := uc.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	metrics.RecordOrderFilled(order.Symbol, string(order.Side))
	if position.Status == domain.PositionStatusClosed {
		metrics.RecordPositionClosed(order.Symbol, string(position.Side), "user")
	}

	logger.Info("position closed/reduced",
		"position_id", position.ID,
		"close_quantity", closeQuantity,
		"pnl", pnl,
		"remaining_quantity", position.Quantity,
	)

	return &PlaceOrderOutput{
		Order:    order,
		Position: position,
		Trade:    trade,
	}, nil
}

func (uc *UseCase) validateInput(input PlaceOrderInput) error {
	if !uc.symbols[input.Symbol] {
		return domain.ErrSymbolNotSupported
	}

	if input.Side != domain.OrderSideBuy && input.Side != domain.OrderSideSell {
		return domain.ErrInvalidOrderSide
	}

	if input.Type != domain.OrderTypeMarket && input.Type != domain.OrderTypeLimit {
		return domain.ErrInvalidOrderType
	}

	if !input.Quantity.IsPositive() {
		return domain.ErrInvalidQuantity
	}

	if !uc.engine.ValidateLeverage(input.Leverage) {
		return domain.ErrInvalidLeverage
	}

	if input.Type == domain.OrderTypeLimit && !input.Price.IsPositive() {
		return domain.ErrInvalidPrice
	}

	return nil
}
