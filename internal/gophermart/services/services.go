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
	GetUserBalance(userID int) (*dto.BalanceResponse, int, error)
	CreateWithdrawal(withdrawal *models.Withdrawal) error
	GetUserWithdrawals(userID int) ([]dto.WithdrawalResponse, int, error)
	WithdrawalExists(orderNumber string) (bool, error)
	GetWithdrawal(req dto.WithdrawRequest, userID int) (int, *models.Withdrawal, error)
}

type OrderService interface {
	UploadOrder(userID int, orderNumber string) (int, error)
	GetOrders(userID int) ([]models.Order, error)
}

type JWTServiceInterface interface {
	GenerateToken(userID int, login string) (string, error)
}

type AccrualClientInterface interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*dto.AccrualResponse, error)
}
