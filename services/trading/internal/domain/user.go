package domain

import (
	"time"
)

type UserID int64

type User struct {
	ID           UserID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
