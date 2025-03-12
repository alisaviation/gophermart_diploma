package repository

import "github.com/rtmelsov/GopherMart/internal/models"

func (r *Repository) GetBalance(id *uint) (*models.DBBalance, *models.Error) {
	return r.db.GetBalance(id)
}

func (r *Repository) AddBalance(id *uint, amount *float64) *models.Error {
	return r.db.AddBalance(id, *amount)
}

func (r *Repository) DeductBalance(id *uint, amount *float64) *models.Error {
	return r.db.DeductBalance(id, *amount)
}
