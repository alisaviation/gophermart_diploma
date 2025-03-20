package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"net/http"
)

func (s *Service) PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error {
	oldOrder, localError := s.repo.GetWithdrawal(withdrawal.OrderNumber)
	if localError == nil {
		if oldOrder.UserID == withdrawal.UserID {
			s.conf.GetLogger().Warn("order belongs to another person")
			return s.ErrorHandler("this order is already assigned by you", http.StatusOK)
		}
		s.ErrorHandler("this order is already assigned to another user", http.StatusConflict)
	}

	localError = utils.PostAccrual(s.conf, withdrawal.OrderNumber)
	if localError != nil {
		return localError
	}

	if oldOrder != nil {
		return localError
	}
	return s.repo.PostOrderWithDraw(withdrawal)
}

func (s *Service) GetWithdrawals(id *uint) (*[]models.Withdrawal, *models.Error) {
	return s.repo.GetWithdrawals(id)
}
