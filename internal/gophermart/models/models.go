package models

import "time"

type User struct {
	ID           int
	Login        string
	PasswordHash string
}

type Order struct {
	ID         int
	UserID     int
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type Balance struct {
	UserID    int
	Current   float64
	Withdrawn float64
}

type Withdrawal struct {
	ID          int
	UserID      int
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
