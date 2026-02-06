package handler

import (
	"net/http"

	"trading/internal/delivery/http/middleware"
	"trading/internal/domain"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo domain.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// UserResponse represents user profile
type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// GetMe returns the current user's profile
// GET /user/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	writeJSON(w, UserResponse{
		ID:        int64(user.ID),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, http.StatusOK)
}
