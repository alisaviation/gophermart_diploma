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
		s.conf.GetLogger().Warn("order is exist")
		if oldOrder.UserID == order.UserID {
			s.conf.GetLogger().Warn("order belongs to another person")
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
	s.conf.GetLogger().Info("try to get bonuses",
		zap.String("AccrualSystemAddress: ", s.conf.GetEnvVariables().AccrualSystemAddress),
		zap.String("DataBaseURL: ", s.conf.GetEnvVariables().DataBaseURL),
		zap.String("RunAddress: ", s.conf.GetEnvVariables().RunAddress),
	)
	orderWithStatus, localError = utils.GetAccrual(s.conf, order.Number)

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
