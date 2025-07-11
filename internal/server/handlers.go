package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/middleware"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// validateOrderNumber проверяет номер заказа с помощью алгоритма Луна
func validateOrderNumber(number string) bool {
	// Проверяем, что номер состоит только из цифр
	if !regexp.MustCompile(`^\d+$`).MatchString(number) {
		return false
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Проверка, на существования пользователя
	existingUser, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		log.Printf("Failed to get user by login: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "Login already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user, err := h.storage.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateJWT(user.ID, user.Login)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	user, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		log.Printf("Failed to get user by login: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
	if !validateOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем, существует ли заказ
	existingOrder, err := h.storage.GetOrderByNumber(r.Context(), orderNumber)
	if err != nil {
		log.Printf("Failed to get order by number: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		log.Printf("Failed to create order: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrdersHandler возвращает список заказов пользователя
func (h *Handlers) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	orders, err := h.storage.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get orders by user ID: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	balance, err := h.storage.GetBalance(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get balance: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req models.WithdrawRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if err = validate.Struct(req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Выполняем списание в транзакции
	_, err = h.storage.ProcessWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		log.Printf("Failed to process withdrawal: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawalsHandler возвращает список списаний пользователя
func (h *Handlers) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.storage.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get withdrawals by user ID: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
