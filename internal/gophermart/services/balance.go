package services

import (
	"fmt"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/gophermart/models"
)

type BalancesService struct {
	Balance database.Balance
}

func NewBalanceService(balance database.Balance) BalanceService {
	return &BalancesService{Balance: balance}
}

func (s *BalancesService) GetUserBalance(userID int) (*models.Balance, error) {
	balance, err := s.Balance.GetBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (s *BalancesService) CreateWithdrawal(withdrawal *models.Withdrawal) error {
	exists, err := s.Balance.WithdrawalExists(withdrawal.OrderNumber)
	if err != nil {
		return fmt.Errorf("failed to check withdrawal existence: %w", err)
	}
	if exists {
		return fmt.Errorf("withdrawal for order %s already exists", withdrawal.OrderNumber)
	}

	if err := s.Balance.CreateWithdrawal(withdrawal); err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

func (s *BalancesService) GetUserWithdrawals(userID int) ([]models.Withdrawal, error) {
	withdrawals, err := s.Balance.GetWithdrawals(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	return withdrawals, nil
}

func (s *BalancesService) WithdrawalExists(orderNumber string) (bool, error) {
	return s.Balance.WithdrawalExists(orderNumber)
}
