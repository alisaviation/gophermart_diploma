package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/valinurovdenis/gomart/internal/app/accrualorder"
	"github.com/valinurovdenis/gomart/internal/app/auth"
	"github.com/valinurovdenis/gomart/internal/app/gzip"
	"github.com/valinurovdenis/gomart/internal/app/logger"
	"github.com/valinurovdenis/gomart/internal/app/orderstorage"
	"github.com/valinurovdenis/gomart/internal/app/service"
	"github.com/valinurovdenis/gomart/internal/app/withdrawstorage"
)

type MartHandler struct {
	Service service.OrderService
	Auth    auth.JwtAuthenticator
}

func (h *MartHandler) AddUserOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("Login")

	number, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.Service.AddUserOrder(r.Context(), login, string(number))

	if errors.Is(err, orderstorage.ErrOrderExists) {
		http.Error(w, err.Error(), http.StatusConflict)
	} else if errors.Is(err, orderstorage.ErrAlreadySent) {
		w.WriteHeader(http.StatusOK)
	} else if errors.Is(err, service.ErrInvalidOrder) || errors.Is(err, accrualorder.ErrNoSuchOrder) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *MartHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("Login")

	orders, err := h.Service.GetUserOrders(r.Context(), login)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *MartHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("Login")

	userBalance, err := h.Service.GetUserBalance(r.Context(), login)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userBalance)
}

func (h *MartHandler) WithdrawOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("Login")

	var withdraw withdrawstorage.UserWithdraw
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	withdraw.Login = login

	err := h.Service.AddUserWithdraw(r.Context(), withdraw)

	if errors.Is(err, service.ErrNotEnoughBalance) {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	} else if errors.Is(err, service.ErrInvalidOrder) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MartHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("Login")

	withdrawals, err := h.Service.GetUserWithdrawals(r.Context(), login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}

func NewMartHandler(service service.OrderService, auth auth.JwtAuthenticator) *MartHandler {
	return &MartHandler{Service: service, Auth: auth}
}

func MartRouter(handler MartHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(gzip.GzipMiddleware)

	r.Post("/api/user/register", handler.Auth.Register)
	r.Post("/api/user/login", handler.Auth.Login)

	r.Route("/", func(r chi.Router) {
		r.Use(handler.Auth.Authenticate)
		r.Post("/api/user/orders", handler.AddUserOrder)
		r.Get("/api/user/orders", handler.GetUserOrders)
		r.Get("/api/user/balance", handler.GetUserBalance)
		r.Post("/api/user/balance/withdraw", handler.WithdrawOrder)
		r.Get("/api/user/withdrawals", handler.GetWithdrawals)
	})

	return r
}
