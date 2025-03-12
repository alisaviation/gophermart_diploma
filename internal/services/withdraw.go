package services

import "github.com/rtmelsov/GopherMart/internal/models"

func (s *Service) PostBalanceWithdraw(order *models.DBWithdrawal) *models.Error {
	return s.repo.PostBalanceWithdraw(order)
}

func (s *Service) GetWithdrawals(id *uint) (*[]models.DBWithdrawal, *models.Error) {
	return s.repo.GetWithdrawals(id)
}
