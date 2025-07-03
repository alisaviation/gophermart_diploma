package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alisaviation/internal/gophermart/models"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresDatabase(db *sql.DB) (*PostgresStorage, error) {
	storage := &PostgresStorage{db: db}
	if err := storage.createTable(context.Background(), db); err != nil {
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}
	return storage, nil
}

func (p *PostgresStorage) createTable(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
		    id SERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
		    id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			number TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL,
			accrual DECIMAL(10, 2) DEFAULT 0,
		    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to orders users table: %w", err)
	}

	return tx.Commit()
}

func (p *PostgresStorage) CreateUser(user models.User) error {
	_, err := p.db.Exec(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		user.Login, user.PasswordHash,
	)
	return err
}

func (p *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow(
		"SELECT id, login, password_hash FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *PostgresStorage) CreateOrder(order *models.Order) error {
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) 
              VALUES ($1, $2, $3, $4)`
	_, err := p.db.Exec(query, order.UserID, order.Number, order.Status, order.UploadedAt)
	return err
}

func (p *PostgresStorage) GetOrderByNumber(number string) (*models.Order, error) {
	var order models.Order
	query := `SELECT id, user_id, number, status, uploaded_at FROM orders WHERE number = $1`
	err := p.db.QueryRow(query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &order, nil
}
