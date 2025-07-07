package handlers

import (
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
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	var goods []dto.AccrualGood
	var orderNumber string

	contentType := r.Header.Get("Content-Type")
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}

	//body, err := io.ReadAll(r.Body)
	//if err != nil {
	//	logger.Log.Info("Failed to read request body")
	//	http.Error(w, "Bad request", http.StatusBadRequest)
	//	return
	//}
	//defer r.Body.Close()
	//
	//orderNumber := string(body)
	//if orderNumber == "" {
	//	http.Error(w, "Empty order number", http.StatusBadRequest)
	//	return
	//}

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

	default:
		http.Error(w, "Unsupported content type", http.StatusBadRequest)
		return
	}
	if orderNumber == "" {
		http.Error(w, "Empty order number", http.StatusBadRequest)
		return
	}
	status, err := h.orderService.UploadOrder(userID, orderNumber, goods)
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
	// Преобразуем в DTO для ответа
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
