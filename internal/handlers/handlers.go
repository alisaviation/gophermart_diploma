package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/middleware"
	"github.com/alisaviation/pkg/logger"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.authService.Register(req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrLoginTaken):
			respondWithError(w, http.StatusConflict, "Login already taken")
		default:
			logger.Log.Error("Registration failed", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	respondWithToken(w, http.StatusOK, "User registered successfully", token)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.authService.Login(req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			respondWithError(w, http.StatusUnauthorized, "Invalid login or password")
		default:
			logger.Log.Error("Login failed", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	respondWithToken(w, http.StatusOK, "Successfully authenticated", token)
}

type OrderHandler struct {
	orderService services.OrderService
	serverCtx    context.Context
}

func NewOrderHandler(orderService services.OrderService, serverCtx context.Context) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		serverCtx:    serverCtx,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	var goods []dto.AccrualGood
	var orderNumber string
	ctx, cancel := context.WithTimeout(h.serverCtx, 10*time.Second)
	defer cancel()

	contentType := r.Header.Get("Content-Type")
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	switch contentType {
	case "text/plain":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		orderNumber = string(body)
		goods = []dto.AccrualGood{}

	case "application/json":
		var req struct {
			Order string            `json:"order"`
			Goods []dto.AccrualGood `json:"goods"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		orderNumber = req.Order
		goods = req.Goods

		// Валидация товаров
		for _, good := range goods {
			if good.Description == "" || good.Price <= 0 {
				logger.Log.Warn("Invalid good data",
					zap.Any("good", good),
					zap.String("order", orderNumber))
				http.Error(w, "Invalid good data", http.StatusBadRequest)
				return
			}
		}

	default:
		http.Error(w, "Unsupported content type", http.StatusBadRequest)
		return
	}
	if orderNumber == "" {
		http.Error(w, "Empty order number", http.StatusBadRequest)
		return
	}
	status, err := h.orderService.UploadOrder(ctx, userID, orderNumber, goods)
	if err != nil {
		logger.Log.Error("Failed to process order",
			zap.Error(err),
			zap.String("orderNumber", orderNumber),
			zap.Int("userID", userID))

		switch status {
		case http.StatusBadRequest:
			http.Error(w, "Invalid order number format", status)
		case http.StatusUnprocessableEntity:
			http.Error(w, "Invalid order number", status)
		case http.StatusConflict:
			http.Error(w, "Order already uploaded by another user", status)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(status)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		logger.Log.Error("UserID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetOrders(userID)
	if err != nil {
		logger.Log.Error("Failed to get user orders",
			zap.Error(err),
			zap.Int("userID", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	type orderResponse struct {
		Number     string    `json:"number"`
		Status     string    `json:"status"`
		Accrual    float64   `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		resp := orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		if order.Status == "PROCESSED" {
			resp.Accrual = order.Accrual
		}

		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		logger.Log.Error("Failed to encode response",
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

type BalanceHandler struct {
	balanceService services.BalanceService
}

func NewBalanceHandler(balanceService services.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		logger.Log.Error("User ID not found in context")
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		logger.Log.Error("Failed to encode balance response",
			zap.Int("userID", userID),
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
