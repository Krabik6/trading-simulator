package account

import (
	"context"

	"trading/internal/domain"
)

type UseCase struct {
	accountRepo  domain.AccountRepository
	positionRepo domain.PositionRepository
}

func NewUseCase(accountRepo domain.AccountRepository, positionRepo domain.PositionRepository) *UseCase {
	return &UseCase{
		accountRepo:  accountRepo,
		positionRepo: positionRepo,
	}
}

type AccountInfo struct {
	Balance         string `json:"balance"`
	Equity          string `json:"equity"`
	UsedMargin      string `json:"used_margin"`
	AvailableMargin string `json:"available_margin"`
	UnrealizedPnL   string `json:"unrealized_pnl"`
	MarginRatio     string `json:"margin_ratio"`
}

func (uc *UseCase) GetAccountInfo(ctx context.Context, userID domain.UserID) (*AccountInfo, error) {
	account, err := uc.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	positions, err := uc.positionRepo.GetOpenByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := account.CalculateSummary(positions)

	return &AccountInfo{
		Balance:         summary.Balance.StringFixed(2),
		Equity:          summary.Equity.StringFixed(2),
		UsedMargin:      summary.UsedMargin.StringFixed(2),
		AvailableMargin: summary.AvailableMargin.StringFixed(2),
		UnrealizedPnL:   summary.UnrealizedPnL.StringFixed(2),
		MarginRatio:     summary.MarginRatio.StringFixed(4),
	}, nil
}
