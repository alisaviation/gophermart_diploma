package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (db *DB) PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error {
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
	if err := tx.First(&user, withdrawal.UserID).Error; err != nil {
		tx.Rollback()
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	// так как данные по балансу не вложение
	// меняем в объекте клиента данные баланса
	db.conf.GetLogger().Info("to change balance in postorderwithdraw", zap.Float64("user balance", user.Current), zap.Float64("sum to subtract", withdrawal.Sum))
	user.Current -= withdrawal.Sum
	if user.Current < 0 {
		tx.Rollback()
		return db.ErrorHandler(
			"",
			http.StatusPaymentRequired,
		)
	}
	user.Withdraw += withdrawal.Sum

	// дальше сохранение данных по балансу в таблице клиента
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	// сохранение списка вычитания в таблице withdrawals
	withdrawal.ProcessedAt = time.Now()
	if err := tx.Create(withdrawal).Error; err != nil {
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

func (db *DB) GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error) {
	var user *models.DBUser
	result := db.db.Preload("Withdrawals").First(&user, id)
	if result.Error != nil {
		return nil, db.ErrorHandler(
			result.Error.Error(),
			http.StatusInternalServerError,
		)
	}

	return &user.Withdrawals, nil
}

func (db *DB) GetWithdrawal(orderNumber string) (*models.DBWithdrawal, *models.Error) {
	var withdrawal models.DBWithdrawal

	// Ищем конкретный заказ по номеру
	err := db.db.Where("order_number = ?", orderNumber).First(&withdrawal).Error
	if err != nil {
		return nil, db.ErrorHandler(
			err.Error(),
			http.StatusInternalServerError,
		)
	}

	return &withdrawal, nil
}
