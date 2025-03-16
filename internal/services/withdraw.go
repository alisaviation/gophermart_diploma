package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
)

func (s *Service) PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error {
	oldOrder, localError := s.repo.GetWithdrawal(withdrawal.OrderNumber)
	if localError == nil {
		if oldOrder.UserID == withdrawal.UserID {
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
	return s.repo.PostOrderWithDraw(withdrawal)
}

func (s *Service) GetWithdrawals(id *uint) (*[]models.Withdrawal, *models.Error) {
	return s.repo.GetWithdrawals(id)
}
