package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

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
	//_, err = tx.ExecContext(ctx, `
	//    CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)
	//`)
	//if err != nil {
	//	return fmt.Errorf("failed to create orders user_id index: %w", err)
	//}
	//
	//_, err = tx.ExecContext(ctx, `
	//    CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id)
	//`)
	//if err != nil {
	//	return fmt.Errorf("failed to create withdrawals user_id index: %w", err)
	//}
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
	query := `INSERT INTO orders (user_id, number, status, accrual, uploaded_at) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := p.db.Exec(query, order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt)
	return err
}

func (p *PostgresStorage) UpdateOrderStatus(number string, status string) error {
	_, err := p.db.Exec(
		"UPDATE orders SET status = $1 WHERE number = $2",
		status, number,
	)
	return err
}

func (p *PostgresStorage) UpdateOrderFromAccrual(number string, status string, accrual float64) error {
	query := `
        UPDATE orders 
        SET status = $1, accrual = $2 
        WHERE number = $3`

	_, err := p.db.Exec(query, status, accrual, number)
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

func (p *PostgresStorage) GetOrdersByUser(userID int) ([]models.Order, error) {
	query := `
        SELECT id, user_id, number, status, accrual, uploaded_at 
        FROM orders 
        WHERE user_id = $1 
        ORDER BY uploaded_at DESC`

	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *PostgresStorage) GetOrdersByStatuses(statuses []string) ([]models.Order, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	query := `
        SELECT id, user_id, number, status, accrual, uploaded_at 
        FROM orders 
        WHERE status = ANY($1)
        ORDER BY uploaded_at DESC`

	rows, err := p.db.Query(query, pq.Array(statuses))
	if err != nil {
		return nil, fmt.Errorf("failed to query orders by statuses: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (p *PostgresStorage) GetBalance(userID int) (*models.Balance, error) {
	balance := &models.Balance{
		UserID: userID,
	}

	err := p.db.QueryRow(`
        SELECT COALESCE(
            (SELECT SUM(accrual) 
             FROM orders 
             WHERE user_id = $1 AND status = 'PROCESSED')
            -
            (SELECT COALESCE(SUM(sum), 0)
             FROM withdrawals 
             WHERE user_id = $1)
        , 0) AS current_balance`,
		userID).Scan(&balance.Current)

	if err != nil {
		return nil, fmt.Errorf("failed to get current balance:: %w", err)
	}

	err = p.db.QueryRow(`
        SELECT COALESCE(SUM(sum), 0) 
        FROM withdrawals 
        WHERE user_id = $1`,
		userID).Scan(&balance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawn balance: %w", err)
	}

	return balance, nil
}
