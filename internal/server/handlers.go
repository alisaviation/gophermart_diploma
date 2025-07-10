package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/middleware"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/utils"
)

// Handlers содержит все HTTP обработчики
type Handlers struct {
	storage        storage.Storage
	authService    *services.AuthService
	accrualService *services.AccrualService
}

// NewHandlers создает новые обработчики
func NewHandlers(storage storage.Storage, authService *services.AuthService, accrualService *services.AccrualService) *Handlers {
	return &Handlers{
		storage:        storage,
		authService:    authService,
		accrualService: accrualService,
	}
}

// RegisterHandler обрабатывает регистрацию пользователя
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Валидация
	if !utils.ValidateLogin(req.Login) || !utils.ValidatePassword(req.Password) {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Проверка, на существования пользователя
	existingUser, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "Login already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user, err := h.storage.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateJWT(user.ID, user.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LoginHandler обрабатывает аутентификацию пользователя
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Валидация
	if !utils.ValidateLogin(req.Login) || !utils.ValidatePassword(req.Password) {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	user, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := h.authService.CheckPassword(user.Password, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateJWT(user.ID, user.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// UploadOrderHandler обрабатывает загрузку номера заказа
func (h *Handlers) UploadOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)

	// Валидация номера заказа
	if !utils.ValidateOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем, существует ли заказ
	existingOrder, err := h.storage.GetOrderByNumber(r.Context(), orderNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			// Заказ уже загружен этим пользователем
			w.WriteHeader(http.StatusOK)
			return
		} else {
			// Заказ уже загружен другим пользователем
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			return
		}
	}

	// Создаем новый заказ
	_, err = h.storage.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrdersHandler возвращает список заказов пользователя
func (h *Handlers) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.storage.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразуем в формат ответа
	responses := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = models.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetBalanceHandler возвращает баланс пользователя
func (h *Handlers) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.storage.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WithdrawHandler обрабатывает списание средств
func (h *Handlers) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Валидация номера заказа
	if !utils.ValidateOrderNumber(req.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем баланс
	balance, err := h.storage.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if balance.Current < req.Sum {
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	// Создаем списание
	_, err = h.storage.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Обновляем баланс
	newCurrent := balance.Current - req.Sum
	newWithdrawn := balance.Withdrawn + req.Sum
	err = h.storage.UpdateBalance(r.Context(), userID, newCurrent, newWithdrawn)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawalsHandler возвращает список списаний пользователя
func (h *Handlers) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.storage.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразуем в формат ответа
	responses := make([]models.WithdrawalResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		responses[i] = models.WithdrawalResponse{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
