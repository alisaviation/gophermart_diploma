package database

import (
	"github.com/alisaviation/internal/gophermart/models"
)

type Storage interface {
	UserRepository
	//OrderRepository
	//BalanceRepository
}

type UserRepository interface {
	CreateUser(user models.User) error
	GetUserByLogin(login string) (*models.User, error)
}

type OrderRepository interface {
	CreateOrder(userID int, number string) error
	GetOrdersByUser(userID int) ([]models.Order, error)
	GetOrderByNumber(number string) (*models.Order, error)
}

type BalanceRepository interface {
	GetBalance(userID int) (*models.Balance, error)
	UpdateBalance(userID int, current, withdrawn float64) error
	CreateWithdrawal(userID int, orderNumber string, sum float64) error
	GetWithdrawals(userID int) ([]models.Withdrawal, error)
}
