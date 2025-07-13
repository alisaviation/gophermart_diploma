package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/services"
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
