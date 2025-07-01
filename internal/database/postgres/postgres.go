package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alisaviation/internal/gophermart/models"
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

func (r *PostgresStorage) createTable(ctx context.Context, db *sql.DB) error {
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

	return tx.Commit()
}

func (r *PostgresStorage) CreateUser(user models.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		user.Login, user.PasswordHash,
	)
	return err
}

func (r *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
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
