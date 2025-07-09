package db

import (
	"database/sql"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepoPG struct {
	db *sql.DB
}

func NewOrderRepoPG(db *sql.DB) *OrderRepoPG {
	return &OrderRepoPG{db: db}
}

func (r *OrderRepoPG) CreateOrder(orderNumber string, userID int64) error {
	_, err := r.db.Exec(`INSERT INTO orders (order_number, user_id) VALUES ($1, $2)`, orderNumber, userID)
	return err
}

func (r *OrderRepoPG) GetOrderByNumber(orderNumber string) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, created_at FROM orders WHERE order_number=$1`, orderNumber).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrderByNumberAndUserID(orderNumber string, userID int64) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, created_at FROM orders WHERE order_number=$1 AND user_id=$2`, orderNumber, userID).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
