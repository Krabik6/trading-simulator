package domain

import "errors"

var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Account errors
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInsufficientMargin  = errors.New("insufficient margin")

	// Order errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderNotPending    = errors.New("order is not pending")
	ErrInvalidOrderSide   = errors.New("invalid order side")
	ErrInvalidOrderType   = errors.New("invalid order type")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrInvalidLeverage    = errors.New("invalid leverage")
	ErrInvalidPrice       = errors.New("invalid price")
	ErrSymbolNotSupported = errors.New("symbol not supported")

	// Position errors
	ErrPositionNotFound      = errors.New("position not found")
	ErrPositionNotOpen       = errors.New("position is not open")
	ErrPositionAlreadyExists = errors.New("position already exists for this symbol")
	ErrInvalidStopLoss       = errors.New("invalid stop loss")
	ErrInvalidTakeProfit     = errors.New("invalid take profit")
	ErrInvalidClosePercent   = errors.New("close percent must be between 1 and 100")

	// Trade errors
	ErrTradeNotFound = errors.New("trade not found")

	// Price errors
	ErrPriceNotAvailable = errors.New("price not available")
)
