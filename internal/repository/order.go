package repository

import "github.com/rtmelsov/GopherMart/internal/models"

func (r *Repository) PostOrders(order *models.DBOrder) *models.Error {
	r.db.PostOrders(order)
	return nil
}

func (r *Repository) GetOrders(id *uint) (*[]models.DBOrder, *models.Error) {
	return r.db.GetOrders(id)
}

func (r *Repository) GetOrder(id *uint, orderNumber int64) (*models.DBOrder, *models.Error) {
	return r.db.GetOrder(id, orderNumber)
}
