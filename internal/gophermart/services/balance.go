package services

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/pkg/logger"
)

type BalancesService struct {
	Balance database.Balance
}

func NewBalanceService(balance database.Balance) BalanceService {
	return &BalancesService{Balance: balance}
}

func (s *BalancesService) GetUserBalance(userID int) (*dto.BalanceResponse, int, error) {
	balance, err := s.Balance.GetBalance(userID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get balance: %w", err)
	}

	response := &dto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	return response, http.StatusOK, nil
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

func (s *BalancesService) GetUserWithdrawals(userID int) ([]dto.WithdrawalResponse, int, error) {
	withdrawals, err := s.Balance.GetWithdrawals(userID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	if len(withdrawals) == 0 {
		return nil, http.StatusNoContent, nil
	}

	response := make([]dto.WithdrawalResponse, 0, len(withdrawals))
	for _, wd := range withdrawals {
		response = append(response, dto.WithdrawalResponse{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		})
	}

	return response, http.StatusOK, nil
}

func (s *BalancesService) WithdrawalExists(orderNumber string) (bool, error) {
	return s.Balance.WithdrawalExists(orderNumber)
}

func (s *BalancesService) GetWithdrawal(req dto.WithdrawRequest, userID int) (int, *models.Withdrawal, error) {
	if _, err := strconv.Atoi(req.Order); err != nil {
		return http.StatusUnprocessableEntity, nil, fmt.Errorf("invalid order number format")
	}

	if !ValidateOrderNumber(req.Order) {
		return http.StatusUnprocessableEntity, nil, fmt.Errorf("invalid order number")
	}

	exists, err := s.WithdrawalExists(req.Order)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("failed to check withdrawal existence: %w", err)
	}
	if exists {
		return http.StatusConflict, nil, fmt.Errorf("withdrawal for this order already exists")
	}

	currentBalance, status, err := s.GetUserBalance(userID)
	if err != nil {
		return status, nil, fmt.Errorf("failed to check balance: %w", err)
	}

	if currentBalance.Current < req.Sum {
		logger.Log.Warn("Insufficient funds",
			zap.Int("userID", userID),
			zap.Float64("available", currentBalance.Current),
			zap.Float64("requested", req.Sum))
		return http.StatusPaymentRequired, nil, fmt.Errorf("insufficient funds")
	}

	withdrawal := &models.Withdrawal{
		UserID:      userID,
		OrderNumber: req.Order,
		Sum:         req.Sum,
		ProcessedAt: time.Now(),
	}

	if err := s.CreateWithdrawal(withdrawal); err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("failed to register withdrawal: %w", err)
	}

	return http.StatusOK, withdrawal, nil
}
