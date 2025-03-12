package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
)

func (db *DB) PostBalanceWithdraw(withdraw *models.DBWithdrawal) *models.Error {
	db.db.Create(withdraw)
	return nil
}

func (db *DB) GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error) {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return nil, &models.Error{
			Error: result.Error.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	return &user.Withdrawals, nil
}
