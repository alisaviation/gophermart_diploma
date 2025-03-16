package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"net/http"
)

func (s *Service) PostOrders(order *models.DBOrder) *models.Error {
	oldOrder, localError := s.repo.GetOrder(order.Number)
	if localError == nil {
		if oldOrder.UserID == order.UserID {
			return &models.Error{
				Error: "",
				Code:  http.StatusOK,
			}
		}
		return &models.Error{
			Error: "",
			Code:  http.StatusConflict,
		}
	}
	if oldOrder != nil {
		return localError
	}
	var orderWithStatus *models.Accrual
	orderWithStatus, localError = utils.GetAccrual(s.conf.GetEnvVariables().AccuralSystemAddress, order.Number)

	order.Status = orderWithStatus.Status
	order.Accrual = &orderWithStatus.Accrual
	if localError != nil {
		return localError
	}
	return s.repo.PostOrders(order)
}

func (s *Service) GetOrders(id *uint) (*[]models.Order, *models.Error) {
	return s.repo.GetOrders(id)
}
