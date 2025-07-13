package services

import (
	"context"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/models"
)

type AuthService interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

type BalanceService interface {
	GetUserBalance(userID int) (*models.Balance, error)
	CreateWithdrawal(withdrawal *models.Withdrawal) error
	GetUserWithdrawals(userID int) ([]models.Withdrawal, error)
	WithdrawalExists(orderNumber string) (bool, error)
}

type OrderService interface {
	UploadOrder(ctx context.Context, userID int, orderNumber string, goods []dto.AccrualGood) (int, error)
	GetOrders(userID int) ([]models.Order, error)
	ValidateOrderNumber(number string) bool
}

type JWTServiceInterface interface {
	GenerateToken(userID int, login string) (string, error)
}
