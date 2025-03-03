package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	internalErrors "github.com/ruslantos/go-musthave-diploma-tpl/internal/errors"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/middlware/logger"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
)

type User struct {
	Login    string `db:"login"`
	Password string `db:"password"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) InitStorage() error {
	_, err := r.db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS users(login TEXT,password TEXT);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_login ON users(login);`)
	if err != nil {
		logger.Get().Error("Failed to create storage", zap.Error(err))
		return err
	}

	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, login, password string) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, login, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return internalErrors.ErrLoginAlreadyExists
		}
		logger.Get().Error("Failed to create user", zap.Error(err))
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	var user User
	query := `SELECT login, password FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Get().Error("Failed to get user by login", zap.Error(err))
		return nil, err
	}
	return &user, nil
}
