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
	GetOrders(id *uint) (*[]models.DBOrder, *models.Error)
	GetBalance(id *uint) (*models.DBBalance, *models.Error)
	PostBalanceWithdraw(order *models.DBWithdrawal) *models.Error
	GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error)
	GetOrder(id *uint, orderNumber int64) (*models.DBOrder, *models.Error)
	AddBalance(id *uint, amount *float64) *models.Error
	DeductBalance(id *uint, amount *float64) *models.Error
}

func GetRepository(conf config.ConfigI, db db.DBI) RepositoryI {
	return &Repository{
		conf: conf,
		db:   db,
	}
}
