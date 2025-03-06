package storage

import (
	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

type StorageInterface interface {
	Init() error

	RegisterNewUser(user add.User) error

	AddNewOrder(user add.User, orderNumber int) error

	CheckUserLogin(login string) error

	CheckUser(user add.User) (bool, error)

	GetUserBalance(login string) (add.Balance, error)

	GetAllOrders(orders *[]add.Order, login string) error
}
