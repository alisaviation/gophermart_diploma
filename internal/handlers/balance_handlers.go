package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/models"
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

	balance, err := h.balanceService.GetUserBalance(userID)
	if err != nil {
		logger.Log.Error("Failed to get user balance",
			zap.Int("userID", userID),
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := dto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	writeJSONResponse(w, http.StatusOK, response, zap.Int("userID", userID))
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

	if _, err := strconv.Atoi(req.Order); err != nil {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	if !h.orderService.ValidateOrderNumber(req.Order) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	exists, err := h.balanceService.WithdrawalExists(req.Order)
	if err != nil {
		http.Error(w, "Failed to check withdrawal existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Withdrawal for this order already existsr", http.StatusConflict)
		return
	}

	currentBalance, err := h.balanceService.GetUserBalance(userID)
	if err != nil {
		http.Error(w, "Failed to check balance", http.StatusInternalServerError)
		return
	}

	if currentBalance.Current < req.Sum {
		logger.Log.Warn("Insufficient funds",
			zap.Int("userID", userID),
			zap.Float64("available", currentBalance.Current),
			zap.Float64("requested", req.Sum))
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	withdrawal := models.Withdrawal{
		UserID:      userID,
		OrderNumber: req.Order,
		Sum:         req.Sum,
		ProcessedAt: time.Now(),
	}

	if err := h.balanceService.CreateWithdrawal(&withdrawal); err != nil {
		http.Error(w, "Failed to register withdrawal", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, nil, zap.Int("userID", userID), zap.String("order", req.Order), zap.Float64("sum", req.Sum))
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetUserWithdrawals(userID)
	if err != nil {
		http.Error(w, "Failed to get withdrawals", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		http.Error(w, "No content", http.StatusNoContent)
		return
	}

	response := make([]dto.WithdrawalResponse, 0, len(withdrawals))
	for _, wd := range withdrawals {
		response = append(response, dto.WithdrawalResponse{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		})
	}

	writeJSONResponse(w, http.StatusOK, response, zap.Int("userID", userID))

}
