package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS withdrawals (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id),
            order_number TEXT NOT NULL UNIQUE,
            sum DECIMAL(10, 2) NOT NULL,
            processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)

	if err != nil {
		return fmt.Errorf("failed to create withdrawals table: %w", err)
	}
	return tx.Commit()
}
