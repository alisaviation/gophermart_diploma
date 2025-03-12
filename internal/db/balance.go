package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
)

func (db *DB) AddBalance(id *uint, amount float64) *models.Error {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return &models.Error{
			Error: result.Error.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	newBalance := models.DBBalance{
		UserID:  user.Balance.UserID,
		Current: user.Balance.Current - amount,
	}

	user.Balance = newBalance
	db.db.Save(user)

	return nil
}

func (db *DB) DeductBalance(id *uint, amount float64) *models.Error {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return &models.Error{
			Error: result.Error.Error(),
			Code:  http.StatusInternalServerError,
		}
	}
	newBalance := models.DBBalance{
		UserID:   user.Balance.UserID,
		Current:  user.Balance.Current - amount,
		Withdraw: user.Balance.Withdraw + amount,
	}

	user.Balance = newBalance
	db.db.Save(user)

	return nil
}

func (db *DB) GetBalance(id *uint) (*models.DBBalance, *models.Error) {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return nil, &models.Error{
			Error: result.Error.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	return &user.Balance, nil
}
