package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/models"
)

type MockBalance struct {
	mock.Mock
}

func (m *MockBalance) GetBalance(userID int) (*models.Balance, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Balance), args.Error(1)
}

func (m *MockBalance) CreateWithdrawal(withdrawal *models.Withdrawal) error {
	args := m.Called(withdrawal)
	return args.Error(0)
}

func (m *MockBalance) GetWithdrawals(userID int) ([]models.Withdrawal, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

func (m *MockBalance) WithdrawalExists(orderNumber string) (bool, error) {
	args := m.Called(orderNumber)
	return args.Bool(0), args.Error(1)
}
