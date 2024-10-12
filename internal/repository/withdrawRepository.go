package repository

import (
	"github.com/shopspring/decimal"
	"gophermart/internal/interfaces"
	"gophermart/storage"
	"time"
)

type WithdrawRepository struct {
	DBStorage *storage.PgStorage
}

func (wr *WithdrawRepository) GetDBStorage() interfaces.DBStorageInterface {
	return wr.DBStorage
}

func (wr *WithdrawRepository) Withdrawals(userID int) ([]interfaces.WithdrawInfo, error) {
	var withdrawalInfoArray []interfaces.WithdrawInfo

	query := "SELECT order_number, sum, created_at FROM withdrawal WHERE user_id = $1 ORDER BY created_at DESC"
	rows, err := wr.DBStorage.Conn.Query(wr.DBStorage.Ctx, query, userID)

	if err != nil {
		return withdrawalInfoArray, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawalInfo interfaces.WithdrawInfo
		if err := rows.Scan(&withdrawalInfo.OrderNumber, &withdrawalInfo.Sum, &withdrawalInfo.ProcessedAt); err != nil {
			return withdrawalInfoArray, err
		}

		withdrawalInfoArray = append(withdrawalInfoArray, withdrawalInfo)
	}

	if err := rows.Err(); err != nil {
		return withdrawalInfoArray, err
	}

	return withdrawalInfoArray, nil
}

func (wr *WithdrawRepository) GetCurrentBalance(userID int) (decimal.Decimal, error) {
	var userBalance decimal.Decimal

	query := "SELECT current FROM user_balance WHERE user_id = $1"
	if err := wr.DBStorage.Conn.QueryRow(wr.DBStorage.Ctx, query, userID).Scan(&userBalance); err != nil {
		return decimal.Decimal{}, err
	}

	return userBalance, nil
}

func (wr *WithdrawRepository) UpdateUserBalance(userID int, sum decimal.Decimal) error {
	query := "UPDATE user_balance SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2"
	if _, err := wr.DBStorage.Conn.Exec(wr.DBStorage.Ctx, query, sum, userID); err != nil {
		return err
	}
	return nil
}

func (wr *WithdrawRepository) SaveWithdrawal(userID int, orderNumber string, sum decimal.Decimal) error {
	currentTime := time.Now()
	query := "INSERT INTO withdrawal (user_id, order_number, sum, created_at) VALUES ($1, $2, $3, $4)"
	if _, err := wr.DBStorage.Conn.Exec(wr.DBStorage.Ctx, query, userID, orderNumber, sum, currentTime); err != nil {
		return err
	}
	return nil
}
