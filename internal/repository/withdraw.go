package repository

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"time"
)

func (r *Repository) PostOrderWithDraw(withdrawal *models.DBWithdrawal) *models.Error {
	return r.db.PostOrderWithDraw(withdrawal)
}

func (r *Repository) GetWithdrawals(id *uint) (*[]models.Withdrawal, *models.Error) {
	dbWithdrawals, localError := r.db.GetWithdrawals(id)
	if localError != nil {
		return nil, localError
	}
	// Преобразуем `DBWithdrawal` в `Withdrawal`, если модели разные
	withdrawals := make([]models.Withdrawal, len(*dbWithdrawals))
	for i, w := range *dbWithdrawals {
		withdrawals[i] = models.Withdrawal{
			Number:      w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
		}
	}
	return &withdrawals, nil
}

func (r *Repository) GetWithdrawal(orderNumber string) (*models.DBWithdrawal, *models.Error) {
	return r.db.GetWithdrawal(orderNumber)
}
