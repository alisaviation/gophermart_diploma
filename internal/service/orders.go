package service

import (
	"github.com/shopspring/decimal"
	"gophermart/internal/interfaces"
	"gophermart/internal/repository"
)

type OrderService struct {
	OrderRepository *repository.OrderRepository
}

func (or *OrderService) GetOrderID(orderNumber string, userID int) (int, error) {
	return or.OrderRepository.GetOrderID(orderNumber, userID)
}

func (or *OrderService) SaveOrder(orderNumber string, userID int) error {
	return or.OrderRepository.SaveOrder(orderNumber, userID)
}

func (or *OrderService) UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error {
	return or.OrderRepository.UpdateOrder(orderNumber, accrual, status)
}

func (or *OrderService) GetUserOrders(userID int) ([]interfaces.OrderData, error) {
	return or.OrderRepository.GetUserOrders(userID)
}

func (or *OrderService) GetOrderRepository() interfaces.OrderRepositoryInterface {
	return or.OrderRepository
}
