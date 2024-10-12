package interfaces

import (
	"github.com/shopspring/decimal"
	"time"
)

type WithdrawRepositoryInterface interface {
	Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error)
	Withdrawals(userID int) ([]WithdrawInfo, error)
}

type WithdrawInfo struct {
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
