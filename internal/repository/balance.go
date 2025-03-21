package repository

import "github.com/rtmelsov/GopherMart/internal/models"

func (r *Repository) GetBalance(id *uint) (*models.Balance, *models.Error) {
	return r.db.GetBalance(id)
}
