package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (db *DB) PostOrders(order *models.DBOrder) *models.Error {
	// начало транзакций
	tx := db.db.Begin()
	if tx.Error != nil {
		return db.ErrorHandler(
			tx.Error.Error(),
			http.StatusInternalServerError,
		)
	}

	// получаем данные клиента
	var user models.DBUser
	if err := tx.First(&user, order.UserID).Error; err != nil {
		tx.Rollback()
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	// так как данные по балансу не вложение
	// меняем в объекте клиента данные баланса
	db.conf.GetLogger().Info("to change balance in postorder", zap.Float64("user balance", user.Current), zap.Float64("sum to add", *order.Accrual))
	user.Current += *order.Accrual

	// дальше сохранение данных по балансу в таблице клиента
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}
	order.UploadedAt = time.Now()

	// сохранение списка вычитания в таблице withdrawals
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	// отправление данных в DB
	if err := tx.Commit().Error; err != nil {
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	return nil
}

func (db *DB) GetOrders(id *uint) (*[]models.DBOrder, *models.Error) {
	var user *models.DBUser
	result := db.db.Preload("Orders").First(&user, id)
	if result.Error != nil {
		return nil, db.ErrorHandler(
			result.Error.Error(),
			http.StatusInternalServerError,
		)
	}

	return &user.Orders, nil
}

func (db *DB) GetOrder(orderNumber string) (*models.DBOrder, *models.Error) {
	var order models.DBOrder

	// Ищем конкретный заказ по номеру
	err := db.db.Where("number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	return &order, nil
}
