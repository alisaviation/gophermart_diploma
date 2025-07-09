package services

import (
	"fmt"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/gophermart/models"
)

type BalanceService interface {
	GetUserBalance(userID int) (*models.Balance, error)
}

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
