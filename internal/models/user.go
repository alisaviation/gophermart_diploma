package models

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	Sub uint    `json:"sub"`
	Exp float64 `json:"exp"`
	jwt.RegisteredClaims
}

type UserResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type Error struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// Work with clients
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     int64     `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accural"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}

type Withdrawal struct {
	Order      int64     `json:"order"`
	Sum        float64   `json:"sum"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Accural struct {
	Order   int64   `json:"order"`
	Status  string  `json:"status"`
	Accural float64 `json:"accural"`
}
