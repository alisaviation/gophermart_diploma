package interfaces

import (
	"github.com/shopspring/decimal"
	"time"
)

type OrderRepositoryInterface interface {
	GetOrderID(orderNumber string, userID int) (int, error)
	SaveOrder(orderNumber string, userID int) error
	UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error
	GetUserOrders(userID int) ([]OrderData, error)
	GetDBStorage() DBStorageInterface
}

type OrderData struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}
