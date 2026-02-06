package order

import (
	"context"

	"github.com/shopspring/decimal"

	"trading/internal/domain"
	"trading/internal/logger"
	"trading/internal/metrics"
)

type UpdateOrderInput struct {
	Price      *decimal.Decimal
	Quantity   *decimal.Decimal
	StopLoss   *decimal.Decimal
	TakeProfit *decimal.Decimal
}

func (uc *UseCase) GetOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if order.UserID != userID {
		return nil, domain.ErrOrderNotFound
	}

	return order, nil
}

func (uc *UseCase) CancelOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Verify ownership
	if order.UserID != userID {
		return domain.ErrOrderNotFound
	}

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return domain.ErrOrderNotPending
	}

	// Update order status
	order.Status = domain.OrderStatusCancelled
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return err
	}

	metrics.RecordOrderCancelled(order.Symbol)
	logger.Info("order cancelled", "order_id", orderID)

	return nil
}

func (uc *UseCase) GetOrders(ctx context.Context, userID domain.UserID, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return uc.orderRepo.GetByUserID(ctx, userID, limit, offset)
}

func (uc *UseCase) GetPendingOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error) {
	return uc.orderRepo.GetPendingByUserID(ctx, userID)
}

func (uc *UseCase) UpdateOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID, input UpdateOrderInput) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, domain.ErrOrderNotFound
	}

	if !order.IsPending() {
		return nil, domain.ErrOrderNotPending
	}

	if input.Price != nil {
		if !input.Price.IsPositive() {
			return nil, domain.ErrInvalidPrice
		}
		order.Price = *input.Price
	}

	if input.Quantity != nil {
		if !input.Quantity.IsPositive() {
			return nil, domain.ErrInvalidQuantity
		}
		order.Quantity = *input.Quantity
	}

	if input.StopLoss != nil {
		order.StopLoss = input.StopLoss
	}

	if input.TakeProfit != nil {
		order.TakeProfit = input.TakeProfit
	}

	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	logger.Info("order updated", "order_id", orderID)

	return order, nil
}
