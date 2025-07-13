package postgres

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/alisaviation/internal/gophermart/models"
)

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
