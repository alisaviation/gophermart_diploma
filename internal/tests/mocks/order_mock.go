package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/models"
)

type MockOrderDB struct {
	mock.Mock
}

func (m *MockOrderDB) GetOrdersByUser(userID int) ([]models.Order, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderDB) GetOrderByNumber(number string) (*models.Order, error) {
	args := m.Called(number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderDB) CreateOrder(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderDB) UpdateOrderFromAccrual(number string, status string, accrual float64) error {
	args := m.Called(number, status, accrual)
	return args.Error(0)
}
