package price

import (
	"context"
	"sync"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
	"trading/internal/engine"
	"trading/internal/kafka"
	"trading/internal/logger"
	"trading/internal/metrics"
	positionuc "trading/internal/usecase/position"
)

type Processor struct {
	positionRepo  domain.PositionRepository
	priceCache    domain.PriceCache
	engine        *engine.Engine
	tradeProducer *kafka.TradeProducer
	positionUC    *positionuc.UseCase

	mu sync.RWMutex
}

func NewProcessor(
	positionRepo domain.PositionRepository,
	priceCache domain.PriceCache,
	eng *engine.Engine,
	tradeProducer *kafka.TradeProducer,
	positionUC *positionuc.UseCase,
) *Processor {
	return &Processor{
		positionRepo:  positionRepo,
		priceCache:    priceCache,
		engine:        eng,
		tradeProducer: tradeProducer,
		positionUC:    positionUC,
	}
}

// ProcessPrice handles a new price update
func (p *Processor) ProcessPrice(ctx context.Context, price *domain.Price) error {
	// Update price cache
	p.priceCache.Set(price.Symbol, price)
	metrics.RecordPriceUpdate(price.Symbol)

	// Get all open positions for this symbol
	positions, err := p.positionRepo.GetOpenBySymbol(ctx, price.Symbol)
	if err != nil {
		logger.Error("failed to get positions", "symbol", price.Symbol, "error", err)
		return err
	}

	if len(positions) == 0 {
		return nil
	}

	// Calculate mark price (mid price)
	markPrice := decimal.NewFromFloat(price.Mid())

	// Process each position
	for i := range positions {
		pos := &positions[i]
		if err := p.processPosition(ctx, pos, markPrice); err != nil {
			logger.Error("failed to process position",
				"position_id", pos.ID,
				"error", err,
			)
		}
	}

	return nil
}

func (p *Processor) processPosition(ctx context.Context, position *domain.Position, markPrice decimal.Decimal) error {
	// Check triggers (liquidation, SL, TP)
	triggers := p.engine.LiquidationCalc.CheckTriggers(position, markPrice)

	if triggers.ShouldLiquidate {
		return p.handleLiquidation(ctx, position, triggers.TriggerPrice)
	}

	if triggers.ShouldStopLoss {
		return p.handleStopLoss(ctx, position)
	}

	if triggers.ShouldTakeProfit {
		return p.handleTakeProfit(ctx, position)
	}

	// Just update PnL
	p.engine.UpdatePositionPnL(position, markPrice)
	return p.positionRepo.UpdatePnL(ctx, position.ID, markPrice, position.UnrealizedPnL)
}

func (p *Processor) handleLiquidation(ctx context.Context, position *domain.Position, liquidationPrice decimal.Decimal) error {
	logger.Warn("liquidating position",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"liquidation_price", liquidationPrice,
	)

	trade, err := p.positionUC.Liquidate(ctx, position, liquidationPrice)
	if err != nil {
		return err
	}

	// Publish trade event
	if p.tradeProducer != nil && trade != nil {
		if err := p.tradeProducer.PublishTrade(ctx, trade); err != nil {
			logger.Error("failed to publish liquidation trade", "error", err)
		}
	}

	return nil
}

func (p *Processor) handleStopLoss(ctx context.Context, position *domain.Position) error {
	logger.Info("triggering stop loss",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"stop_loss", position.StopLoss,
	)

	trade, err := p.positionUC.TriggerStopLoss(ctx, position)
	if err != nil {
		return err
	}

	// Publish trade event
	if p.tradeProducer != nil && trade != nil {
		if err := p.tradeProducer.PublishTrade(ctx, trade); err != nil {
			logger.Error("failed to publish stop loss trade", "error", err)
		}
	}

	return nil
}

func (p *Processor) handleTakeProfit(ctx context.Context, position *domain.Position) error {
	logger.Info("triggering take profit",
		"position_id", position.ID,
		"symbol", position.Symbol,
		"take_profit", position.TakeProfit,
	)

	trade, err := p.positionUC.TriggerTakeProfit(ctx, position)
	if err != nil {
		return err
	}

	// Publish trade event
	if p.tradeProducer != nil && trade != nil {
		if err := p.tradeProducer.PublishTrade(ctx, trade); err != nil {
			logger.Error("failed to publish take profit trade", "error", err)
		}
	}

	return nil
}

// Start begins processing prices from the consumer
func (p *Processor) Start(ctx context.Context, pricesCh <-chan *domain.Price) {
	logger.Info("price processor started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("price processor stopping")
			return
		case price, ok := <-pricesCh:
			if !ok {
				logger.Info("price channel closed")
				return
			}
			if err := p.ProcessPrice(ctx, price); err != nil {
				logger.Error("failed to process price",
					"symbol", price.Symbol,
					"error", err,
				)
			}
		}
	}
}
