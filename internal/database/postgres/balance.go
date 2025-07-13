package postgres

import (
	"fmt"

	"github.com/alisaviation/internal/gophermart/models"
)

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

func (p *PostgresStorage) CreateWithdrawal(withdrawal *models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)`

	_, err := p.db.Exec(query,
		withdrawal.UserID,
		withdrawal.OrderNumber,
		withdrawal.Sum,
		withdrawal.ProcessedAt)

	return err
}

func (p *PostgresStorage) WithdrawalExists(orderNumber string) (bool, error) {
	var exists bool
	err := p.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM withdrawals 
			WHERE order_number = $1
		)`, orderNumber).Scan(&exists)

	return exists, err
}

func (p *PostgresStorage) GetWithdrawals(userID int) ([]models.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at ASC`

	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return withdrawals, nil
}
