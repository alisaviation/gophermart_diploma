package interfaces

import "github.com/shopspring/decimal"

type OrderServiceInterface interface {
	GetOrderID(orderNumber string, userID int) (int, error)
	SaveOrder(orderNumber string, userID int) error
	UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error
	GetUserOrders(userID int) ([]OrderData, error)
	GetOrderRepository() OrderRepositoryInterface
}
