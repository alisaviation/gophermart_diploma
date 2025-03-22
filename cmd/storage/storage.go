package storage

import (
	add "github.com/Tanya1515/gophermarket/cmd/additional"
	"go.uber.org/zap"
)

type StorageInterface interface {
	Init() error

	RegisterNewUser(user add.User) error

	AddNewOrder(login string, orderNumber string) error

	CheckUserLogin(login string) error

	CheckUser(login, password string) (bool, error)

	CheckUserJWT(login string) error

	GetUserBalance(login string) (add.Balance, error)

	GetAllOrders(orders *[]add.Order, login string) error

	GetSpendOrders(orders *[]add.OrderSpend, login string) error

	ProcessPayPoints(order add.OrderSpend, login string) error

	StartProcessingUserOrder(logger zap.SugaredLogger, result chan add.OrderAcc)

	ProcessAccOrder(order add.Order) error
}
