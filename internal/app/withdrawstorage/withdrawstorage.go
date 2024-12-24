package withdrawstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/valinurovdenis/gomart/internal/app/currencybalance"
)

type UserWithdraw struct {
	Login     string
	Number    string                          `json:"order"`
	Withdraw  currencybalance.CurrencyBalance `json:"sum"`
	Processed time.Time                       `json:"processed_at"`
}

//go:generate mockery --name OrderStorage
type WithdrawStorage interface {
	AddUserWithdraw(context context.Context, order UserWithdraw) error

	GetUserWithdrawals(context context.Context, login string) ([]UserWithdraw, error)
}

type DatabaseWithdrawStorage struct {
	DB *sql.DB
}

func (s *DatabaseWithdrawStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE IF NOT EXISTS withdraw("login" TEXT, "number" BIGINT, "withdraw" BIGINT, "processed" TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	tx.Exec(`CREATE INDEX user_withdraw_index ON orders USING btree(login)`)
	return tx.Commit()
}

var ErrOrderExists = errors.New("conflicting order exists")

func (s *DatabaseWithdrawStorage) AddUserWithdraw(ctx context.Context, order UserWithdraw) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT into withdraw (login, number, withdraw) VALUES ($1,$2,$3)",
		order.Login, order.Number, order.Withdraw.Balance); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE users SET balance=balance-$1, withdrawn=withdrawn+$1 WHERE login=$2",
		order.Withdraw.Balance, order.Login); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DatabaseWithdrawStorage) GetUserWithdrawals(ctx context.Context, login string) ([]UserWithdraw, error) {
	var res []UserWithdraw
	rows, err := s.DB.QueryContext(ctx,
		"SELECT login, number, withdraw, processed FROM withdraw WHERE login = $1", login)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		var order UserWithdraw
		err = rows.Scan(&order.Login, &order.Number, &order.Withdraw.Balance, &order.Processed)
		if err != nil {
			return nil, err
		}

		res = append(res, order)
	}

	return res, nil
}

func NewDatabaseWithdrawStorage(db *sql.DB) *DatabaseWithdrawStorage {
	ret := &DatabaseWithdrawStorage{DB: db}
	ret.Init()
	return ret
}
