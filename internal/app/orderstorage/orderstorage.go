package orderstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valinurovdenis/gomart/internal/app/currencybalance"
)

type OrderStatus string

const (
	Processed  OrderStatus = "PROCESSED"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Registered OrderStatus = "REGISTERED"
	New        OrderStatus = "NEW"
)

func IsFinal(status OrderStatus) bool {
	return status == Processed || status == Invalid
}

type UserOrder struct {
	Login    string
	Number   string                          `json:"number"`
	Status   OrderStatus                     `json:"status"`
	Balance  currencybalance.CurrencyBalance `json:"accrual"`
	Uploaded time.Time                       `json:"uploaded_at"`
}

//go:generate mockery --name OrderStorage
type OrderStorage interface {
	AddUserOrder(context context.Context, order UserOrder) error

	GetUserOrders(context context.Context, login string) ([]UserOrder, error)

	UpdateProcessedOrder(ctx context.Context, order UserOrder) error
}

type DatabaseOrderStorage struct {
	DB *sql.DB
}

func (s *DatabaseOrderStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TYPE status AS ENUM ('PROCESSED', 'INVALID', 'PROCESSING', 'NEW')`)
	tx.Exec(`CREATE TABLE IF NOT EXISTS orders("login" TEXT, "number" BIGINT, "status" status, "balance" BIGINT, "uploaded" TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	tx.Exec(`CREATE UNIQUE INDEX orders_index ON orders USING btree(number)`)
	tx.Exec(`CREATE INDEX user_orders_index ON orders USING btree(login)`)
	return tx.Commit()
}

var ErrOrderExists = errors.New("conflicting order exists")
var ErrAlreadySent = errors.New("order has been already sent by user")

func (s *DatabaseOrderStorage) AddUserOrder(ctx context.Context, order UserOrder) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx,
		`INSERT into orders (login,number,status,balance) VALUES ($1,$2,$3,$4)`,
		order.Login, order.Number, order.Status, order.Balance.Balance)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		var login string
		err = s.DB.QueryRowContext(ctx,
			"SELECT login FROM orders WHERE number = $1", order.Number).Scan(&login)
		if err == nil {
			if login == order.Login {
				return ErrAlreadySent
			} else {
				return ErrOrderExists
			}
		}
	}
	if err == nil && order.Status == Processed {
		_, err = tx.ExecContext(ctx, "UPDATE users SET balance=balance+$1 WHERE login=$2",
			order.Balance.Balance, order.Login)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DatabaseOrderStorage) GetUserOrders(ctx context.Context, login string) ([]UserOrder, error) {
	var res []UserOrder
	rows, err := s.DB.QueryContext(ctx,
		"SELECT login, number, status, balance, uploaded FROM orders WHERE login = $1", login)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		var order UserOrder
		err = rows.Scan(&order.Login, &order.Number, &order.Status, &order.Balance.Balance, &order.Uploaded)
		if err != nil {
			return nil, err
		}

		res = append(res, order)
	}

	return res, nil
}

func (s *DatabaseOrderStorage) UpdateProcessedOrder(ctx context.Context, order UserOrder) error {
	if order.Status != Processed {
		return nil
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "UPDATE orders SET status=$1, balance=$2 WHERE number=$3 and status!='PROCESSED'",
		order.Status, order.Balance.Balance, order.Number)
	if err != nil {
		return err
	}
	rowsUpdated, _ := res.RowsAffected()
	if rowsUpdated != 0 {
		tx.ExecContext(ctx, "UPDATE users SET balance=balance+$1 WHERE number=$2",
			order.Balance, order.Login)
	}
	return tx.Commit()
}

func NewDatabaseOrderStorage(db *sql.DB) *DatabaseOrderStorage {
	ret := &DatabaseOrderStorage{DB: db}
	ret.Init()
	return ret
}
