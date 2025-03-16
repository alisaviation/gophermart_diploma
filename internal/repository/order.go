package repository

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"time"
)

func (r *Repository) PostOrders(order *models.DBOrder) *models.Error {
	r.db.PostOrders(order)
	return nil
}

func (r *Repository) GetOrders(id *uint) (*[]models.Order, *models.Error) {
	dbOrders, localError := r.db.GetOrders(id)
	if localError != nil {
		return nil, localError
	}

	// Преобразуем `DBWithdrawal` в `Withdrawal`, если модели разные
	orders := make([]models.Order, len(*dbOrders))
	for i, w := range *dbOrders {
		orders[i] = models.Order{
			Number:     w.Number,
			Status:     w.Status,
			Accrual:    *w.Accrual,
			UploadedAt: w.UploadedAt.Format(time.RFC3339),
		}
	}

	return &orders, nil
}

func (r *Repository) GetOrder(orderNumber string) (*models.DBOrder, *models.Error) {
	return r.db.GetOrder(orderNumber)
}
