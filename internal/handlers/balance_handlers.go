package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/middleware"
	"github.com/alisaviation/pkg/logger"
)

type BalanceHandler struct {
	balanceService services.BalanceService
	orderService   services.OrderService
}

func NewBalanceHandler(balanceService services.BalanceService, orderService services.OrderService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		orderService:   orderService,
	}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, status, err := h.balanceService.GetUserBalance(userID)
	if err != nil {
		logger.Log.Error("Failed to get user balance",
			zap.Int("userID", userID),
			zap.Error(err))
		http.Error(w, err.Error(), status)
		return
	}

	writeJSONResponse(w, status, response, zap.Int("userID", userID))
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode withdrawal request", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	status, _, err := h.balanceService.GetWithdrawal(req, userID)
	if err != nil {
		logger.Log.Error("Failed to process withdrawal",
			zap.Error(err),
			zap.String("orderNumber", req.Order),
			zap.Int("userID", userID))

		http.Error(w, err.Error(), status)
		return
	}

	writeJSONResponse(w, status, nil, zap.Int("userID", userID), zap.String("order", req.Order), zap.Float64("sum", req.Sum))
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, status, err := h.balanceService.GetUserWithdrawals(userID)
	if err != nil {
		logger.Log.Error("Failed to get user withdrawals",
			zap.Error(err),
			zap.Int("userID", userID))
		http.Error(w, err.Error(), status)
		return
	}

	writeJSONResponse(w, http.StatusOK, response, zap.Int("userID", userID))
}
