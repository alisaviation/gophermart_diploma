package models

import (
	"github.com/golang-jwt/jwt/v5"
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

// User Work with clients
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	Number      string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type Accrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
