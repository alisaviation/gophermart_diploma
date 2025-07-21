package handlers

import (
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/middleware"
	"github.com/alisaviation/pkg/logger"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req := dto.UploadOrderRequest{
		OrderNumber: string(body),
	}

	if req.OrderNumber == "" {
		http.Error(w, "Empty order number", http.StatusBadRequest)
		return
	}

	status, err := h.orderService.UploadOrder(userID, req.OrderNumber)
	if err != nil {
		logger.Log.Error("Failed to process order",
			zap.Error(err),
			zap.String("orderNumber", req.OrderNumber),
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetOrders(userID)
	if err != nil {
		http.Error(w, "Failed to get user orders", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		resp := dto.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		if order.Status == "PROCESSED" {
			resp.Accrual = order.Accrual
		}

		response = append(response, resp)
	}
	writeJSONResponse(w, http.StatusOK, response, zap.Int("userID", userID))
}
