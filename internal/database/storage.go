package database

import (
	"github.com/alisaviation/internal/gophermart/models"
)

type Storage interface {
	User
	Order
	//Balance
}

type User interface {
	CreateUser(user models.User) error
	GetUserByLogin(login string) (*models.User, error)
}

type Order interface {
	CreateOrder(order *models.Order) error
	//GetOrdersByUser(userID int) ([]models.Order, error)
	GetOrderByNumber(number string) (*models.Order, error)
}

type Balance interface {
	GetBalance(userID int) (*models.Balance, error)
	UpdateBalance(userID int, current, withdrawn float64) error
	CreateWithdrawal(userID int, orderNumber string, sum float64) error
	GetWithdrawals(userID int) ([]models.Withdrawal, error)
}
