package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"go.uber.org/zap"
	"net/http"
)

func (s *Service) PostOrders(order *models.DBOrder) *models.Error {
	s.conf.GetLogger().Info("try to get order if exist", zap.String("order number: ", order.Number))
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
	s.conf.GetLogger().Info("try to get bonuses", zap.String("AccuralSystemAddress: ", s.conf.GetEnvVariables().AccuralSystemAddress))
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
