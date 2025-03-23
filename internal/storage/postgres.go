package storage

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib" // Импорт драйвера для работы с PostgreSQL
)

var (
	ErrUserNotFound = errors.New("пользователь не найден")
)

func CreatePostgreStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	storage := &PostgresStorage{db: db}

	err = storage.createTables()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PostgresStorage) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login varchar(255) UNIQUE NOT NULL,
		password varchar(255) NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		author_id INT NOT NULL,
		order_num BIGINT UNIQUE NOT NULL,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) CreateUser(ctx context.Context, login string, password string) error {
	_, err := s.db.QueryContext(ctx, `
	INSERT INTO users (login, password)
	VALUES ($1, $2)
	`, login, password)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) IsUserExist(ctx context.Context, login string) (bool, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", login).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PostgresStorage) GetPasswordFromLogin(ctx context.Context, login string) (string, error) {
	var password string
	err := s.db.QueryRowContext(ctx, "SELECT password FROM users WHERE login = $1", login).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return password, nil

}

func (s *PostgresStorage) GetID(ctx context.Context, login string) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", login).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PostgresStorage) IsOrderExists(ctx context.Context, orderID int) (bool, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM orders WHERE order_num = $1", orderID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func (s *PostgresStorage) GetAuthorOrder(ctx context.Context, orderID int) (int, error) {
	var authorID int
	err := s.db.QueryRowContext(ctx, "SELECT author_id FROM orders WHERE order_num = $1", orderID).Scan(&authorID)
	if err != nil {
		return 0, err
	}
	return authorID, nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, authorID int, orderID int) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO orders (author_id, order_num) VALUES ($1, $2)", authorID, orderID)
	if err != nil {
		return err
	}
	return nil
}
