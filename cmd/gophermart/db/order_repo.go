package db

import (
	"database/sql"
	"time"
)

type Order struct {
	ID          int64     `db:"id"`
	OrderNumber string    `db:"order_number"`
	UserID      int64     `db:"user_id"`
	CreatedAt   time.Time `db:"created_at"`
}

type OrderRepo interface {
	CreateOrder(orderNumber string, userID int64) error
	GetOrderByNumber(orderNumber string) (*Order, error)
	GetOrderByNumberAndUserID(orderNumber string, userID int64) (*Order, error)
}

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

func (r *OrderRepoPG) GetOrderByNumber(orderNumber string) (*Order, error) {
	var o Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, created_at FROM orders WHERE order_number=$1`, orderNumber).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrderByNumberAndUserID(orderNumber string, userID int64) (*Order, error) {
	var o Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, created_at FROM orders WHERE order_number=$1 AND user_id=$2`, orderNumber, userID).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
