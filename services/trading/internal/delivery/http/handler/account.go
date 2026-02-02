package handler

import (
	"net/http"

	"trading/internal/delivery/http/middleware"
	accountuc "trading/internal/usecase/account"
)

type AccountHandler struct {
	accountUC *accountuc.UseCase
}

func NewAccountHandler(accountUC *accountuc.UseCase) *AccountHandler {
	return &AccountHandler{accountUC: accountUC}
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	info, err := h.accountUC.GetAccountInfo(r.Context(), userID)
	if err != nil {
		writeError(w, "failed to get account info", http.StatusInternalServerError)
		return
	}

	writeJSON(w, info, http.StatusOK)
}
