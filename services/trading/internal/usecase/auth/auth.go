package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/shopspring/decimal"

	"trading/internal/auth"
	"trading/internal/domain"
)

type UseCase struct {
	userRepo       domain.UserRepository
	accountRepo    domain.AccountRepository
	jwtService     *auth.JWTService
	initialBalance decimal.Decimal
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthOutput struct {
	UserID int64
	Token  string
}

func NewUseCase(
	userRepo domain.UserRepository,
	accountRepo domain.AccountRepository,
	jwtService *auth.JWTService,
	initialBalance float64,
) *UseCase {
	return &UseCase{
		userRepo:       userRepo,
		accountRepo:    accountRepo,
		jwtService:     jwtService,
		initialBalance: decimal.NewFromFloat(initialBalance),
	}
}

func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	// Check if user already exists
	_, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create account with initial balance
	account := &domain.Account{
		UserID:  user.ID,
		Balance: uc.initialBalance,
	}
	if err := uc.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	// Generate token
	token, err := uc.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		UserID: int64(user.ID),
		Token:  token,
	}, nil
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := uc.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		UserID: int64(user.ID),
		Token:  token,
	}, nil
}

func (uc *UseCase) RefreshToken(ctx context.Context, userID domain.UserID) (*AuthOutput, error) {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new token
	token, err := uc.jwtService.GenerateToken(userID)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		UserID: int64(userID),
		Token:  token,
	}, nil
}

func validateEmail(email string) error {
	if len(email) < 3 || len(email) > 255 {
		return errors.New("email must be between 3 and 255 characters")
	}
	if !strings.Contains(email, "@") {
		return errors.New("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if len(password) > 128 {
		return errors.New("password must be at most 128 characters")
	}
	return nil
}
