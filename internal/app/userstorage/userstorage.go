package userstorage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valinurovdenis/gomart/internal/app/currencybalance"
)

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//go:generate mockery --name UserStorage
type UserStorage interface {
	AddUser(context context.Context, user LoginPassword) error

	GetUserPassword(context context.Context, login string) (string, error)
}

type UserBalance struct {
	Current   currencybalance.CurrencyBalance `json:"current"`
	Withdrawn currencybalance.CurrencyBalance `json:"withdrawn"`
}

//go:generate mockery --name UserBalance
type BalanceStorage interface {
	GetBalance(context context.Context, login string) (UserBalance, error)

	AddBalance(context context.Context, login string, addBalance currencybalance.CurrencyBalance) error

	SetBalance(context context.Context, login string, newBalance UserBalance) error
}

type DatabaseUserStorage struct {
	DB *sql.DB
}

func (s *DatabaseUserStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE IF NOT EXISTS users("login" TEXT, "password" TEXT, "balance" BIGINT DEFAULT 0, "withdrawn" BIGINT DEFAULT 0)`)
	tx.Exec(`CREATE UNIQUE INDEX users_index ON users USING btree(login)`)
	return tx.Commit()
}

var ErrLoginExists = errors.New("conflicting login")

func (s *DatabaseUserStorage) AddUser(ctx context.Context, user LoginPassword) error {
	_, err := s.DB.ExecContext(ctx, "INSERT into users (login, password, balance) VALUES ($1, $2, $3)", user.Login, user.Password, 0)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrLoginExists
	}
	return err
}

func (s *DatabaseUserStorage) GetUserPassword(ctx context.Context, login string) (string, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT password FROM users WHERE login = $1", login)
	var password string
	err := row.Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}

func (s *DatabaseUserStorage) GetBalance(ctx context.Context, login string) (UserBalance, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT balance, withdrawn FROM users WHERE login = $1", login)
	var balance UserBalance
	err := row.Scan(&balance.Current.Balance, &balance.Withdrawn.Balance)
	if err != nil {
		return UserBalance{}, err
	}
	return balance, nil
}

func (s *DatabaseUserStorage) AddBalance(ctx context.Context, login string, addBalance currencybalance.CurrencyBalance) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET balance=balance+$1 WHERE login=$2",
		addBalance.Balance, login)
	if err != nil {
		return err
	}
	return nil
}

func (s *DatabaseUserStorage) SetBalance(ctx context.Context, login string, newBalance UserBalance) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET balance=$1, withdrawn=$2 WHERE login=$3",
		newBalance.Current.Balance, newBalance.Withdrawn.Balance, login)
	if err != nil {
		return err
	}
	return nil
}

func NewDatabaseUserStorage(db *sql.DB) *DatabaseUserStorage {
	ret := &DatabaseUserStorage{DB: db}
	ret.Init()
	return ret
}
