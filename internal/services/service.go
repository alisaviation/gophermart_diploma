package services

import (
	"github.com/rtmelsov/GopherMart/internal/config"
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/repository"
)

type Service struct {
	conf config.ConfigI
	repo repository.RepositoryI
}

type ServiceI interface {
	Login(request *models.User) (*models.UserResponse, *models.Error)
	Register(request *models.User) (*models.UserResponse, *models.Error)
	PostOrders(order *models.DBOrder) *models.Error

	PostOrderWithDraw(order *models.DBWithdrawal) *models.Error

	GetWithdrawals(id *uint) (*[]models.Withdrawal, *models.Error)
	GetOrders(id *uint) (*[]models.Order, *models.Error)
	GetBalance(id *uint) (*models.Balance, *models.Error)
}

func NewService(conf config.ConfigI, r repository.RepositoryI) ServiceI {
	return &Service{
		conf: conf,
		repo: r,
	}
}
