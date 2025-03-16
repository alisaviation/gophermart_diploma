package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
)

func (db *DB) GetBalance(id *uint) (*models.Balance, *models.Error) {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return nil, &models.Error{
			Error: result.Error.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	return &models.Balance{
		Current:  user.Current,
		Withdraw: user.Withdraw,
	}, nil
}
