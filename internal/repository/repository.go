package repository

import (
	"github.com/rtmelsov/GopherMart/internal/config"
	"github.com/rtmelsov/GopherMart/internal/db"
	"github.com/rtmelsov/GopherMart/internal/models"
)

type Repository struct {
	conf config.ConfigI
	db   db.DBI
}

type RepositoryI interface {
	Register(value *models.User) (*models.DBUser, *models.Error)
	Login(value *models.User) (*models.DBUser, *models.Error)

	PostOrders(order *models.DBOrder) *models.Error
	GetOrders(id *uint) (*[]models.Order, *models.Error)
	GetOrder(orderNumber string) (*models.DBOrder, *models.Error)
	GetWithdrawal(orderNumber string) (*models.DBWithdrawal, *models.Error)

	GetBalance(id *uint) (*models.Balance, *models.Error)

	GetWithdrawals(id *uint) (*[]models.Withdrawal, *models.Error)
	PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error
}

func GetRepository(conf config.ConfigI, db db.DBI) RepositoryI {
	return &Repository{
		conf: conf,
		db:   db,
	}
}
