package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"trading/internal/domain"
	authuc "trading/internal/usecase/auth"
)

type AuthHandler struct {
	authUC *authuc.UseCase
}

func NewAuthHandler(authUC *authuc.UseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.authUC.Register(r.Context(), authuc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			writeError(w, "user already exists", http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, AuthResponse{
		UserID: output.UserID,
		Token:  output.Token,
	}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.authUC.Login(r.Context(), authuc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeError(w, "login failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, AuthResponse{
		UserID: output.UserID,
		Token:  output.Token,
	}, http.StatusOK)
}
