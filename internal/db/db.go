package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"sync"
)

type DB struct {
	db  *gorm.DB
	log *zap.Logger
	mu  sync.RWMutex
}

type DBI interface {
	Register(value *models.DBUser) (*models.DBUser, *models.Error)
	Login(value *models.DBUser) (*models.DBUser, *models.Error)
	GetUser(userID uint) (*models.DBUser, *models.Error)

	PostOrders(order *models.DBOrder) *models.Error
	GetOrders(id *uint) (*[]models.DBOrder, *models.Error)
	GetOrder(orderNumber string) (*models.DBOrder, *models.Error)

	GetBalance(id *uint) (*models.Balance, *models.Error)

	GetWithdrawal(orderNumber string) (*models.DBWithdrawal, *models.Error)

	GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error)
	PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error
}

func GetDb(log *zap.Logger, dsn string) (DBI, *models.Error) {
	//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, utils.Error(err, http.StatusInternalServerError)
	}

	err = db.Migrator().DropTable(&models.DBUser{}, &models.DBOrder{}, &models.DBWithdrawal{})
	if err != nil {
		return nil, utils.Error(err, http.StatusInternalServerError)
	}
	err = db.AutoMigrate(&models.DBUser{}, &models.DBOrder{}, &models.DBWithdrawal{})
	if err != nil {
		return nil, utils.Error(err, http.StatusInternalServerError)
	}

	return &DB{
		db:  db,
		log: log,
		mu:  sync.RWMutex{},
	}, nil
}
