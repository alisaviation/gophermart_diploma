package models

import (
	"time"
)

// User пользователь системы
type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserRegisterRequest запрос на регистрацию пользователя
type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserLoginRequest запрос на аутентификацию пользователя
type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Order представляет заказ пользователя
type Order struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// OrderResponse ответ с информацией о заказе
type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance представляет баланс пользователя
type Balance struct {
	UserID    int64   `json:"user_id"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// BalanceResponse ответ с информацией о балансе
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal представляет списание средств
type Withdrawal struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// WithdrawalResponse ответ с информацией о списании
type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// WithdrawRequest запрос на списание средств
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// AccrualResponse ответ от системы начисления баллов
type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
