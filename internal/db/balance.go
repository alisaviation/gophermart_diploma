package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"net/http"
)

func (db *DB) GetBalance(id *uint) (*models.Balance, *models.Error) {
	var user *models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return nil, db.ErrorHandler(
			result.Error.Error(),
			http.StatusInternalServerError,
		)
	}

	return &models.Balance{
		Current:  utils.ToFixed(user.Current, 4),
		Withdraw: utils.ToFixed(user.Withdraw, 4),
	}, nil
}
