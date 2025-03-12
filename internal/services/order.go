package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
)

func (s *Service) PostOrders(order *models.DBOrder) *models.Error {
	_, localError := s.repo.GetOrder(&order.UserID, order.Number)
	if localError == nil {
		return localError
	}
	var orderWithStatus *models.Accural
	orderWithStatus, localError = utils.GetAccrual(s.conf.GetEnvVariables().AccuralSystemAddress, order.Number)
	order.Status = orderWithStatus.Status
	if orderWithStatus.Status == "PROCESSED" {
		localError = s.repo.AddBalance(&order.UserID, &orderWithStatus.Accural)
	}
	if localError != nil {
		return localError
	}
	return s.repo.PostOrders(order)
}

func (s *Service) GetOrders(id *uint) (*[]models.DBOrder, *models.Error) {
	return s.repo.GetOrders(id)
}
