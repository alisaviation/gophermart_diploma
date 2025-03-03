package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	internalErrors "github.com/ruslantos/go-musthave-diploma-tpl/internal/errors"
	service "github.com/ruslantos/go-musthave-diploma-tpl/internal/mart"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/middlware/logger"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.service.Register(ctx, req.Login, req.Password); err != nil {
		if err == internalErrors.ErrLoginAlreadyExists {
			http.Error(w, "Login already exists", http.StatusConflict)
			return
		}
		log.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User registered and authenticated successfully"))
}
