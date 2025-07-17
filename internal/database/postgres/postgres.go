package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresDatabase(db *sql.DB) (*PostgresStorage, error) {
	storage := &PostgresStorage{db: db}
	if err := storage.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	return storage, nil
}
func (p *PostgresStorage) runMigrations() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "migrations")

	driver, err := postgres.WithInstance(p.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

//
//func (p *PostgresStorage) createTable(ctx context.Context, db *sql.DB) error {
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//
//	_, err = tx.ExecContext(ctx, `
//		CREATE TABLE IF NOT EXISTS users (
//		    id SERIAL PRIMARY KEY,
//			login TEXT NOT NULL UNIQUE,
//			password_hash TEXT NOT NULL
//		)
//	`)
//	if err != nil {
//		return fmt.Errorf("failed to create users table: %w", err)
//	}
//
//	_, err = tx.ExecContext(ctx, `
//		CREATE TABLE IF NOT EXISTS orders (
//		    id SERIAL PRIMARY KEY,
//			user_id INTEGER NOT NULL REFERENCES users(id),
//			number TEXT NOT NULL UNIQUE,
//			status TEXT NOT NULL,
//			accrual DECIMAL(10, 2) DEFAULT 0,
//		    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
//		)
//	`)
//	if err != nil {
//		return fmt.Errorf("failed to create orders table: %w", err)
//	}
//
//	_, err = tx.ExecContext(ctx, `
//        CREATE TABLE IF NOT EXISTS withdrawals (
//            id SERIAL PRIMARY KEY,
//            user_id INTEGER NOT NULL REFERENCES users(id),
//            order_number TEXT NOT NULL UNIQUE,
//            sum DECIMAL(10, 2) NOT NULL,
//            processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
//        )
//    `)
//
//	if err != nil {
//		return fmt.Errorf("failed to create withdrawals table: %w", err)
//	}
//	return tx.Commit()
//}
