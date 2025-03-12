package repository

import "github.com/rtmelsov/GopherMart/internal/models"

func (r *Repository) PostBalanceWithdraw(withdraw *models.DBWithdrawal) *models.Error {
	r.db.PostBalanceWithdraw(withdraw)
	return nil
}

func (r *Repository) GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error) {
	return r.db.GetWithdrawals(id)
}
