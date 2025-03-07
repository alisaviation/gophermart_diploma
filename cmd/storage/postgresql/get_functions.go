package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "time"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (db *PostgreSQL) CheckUserLogin(login string) error {
	var value string

	row := db.dbConn.QueryRow("SELECT Login FROM Users WHERE Login = $1", login)

	err := row.Scan(&value)
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("user with login %s already exists", login)
	} else if (err != nil) && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	return nil
}

func (db *PostgreSQL) CheckUser(login, password string) (ok bool, err error) {
	ok = true
	row := db.dbConn.QueryRow(`SELECT (password = crypt($1, password)) 
								AS password_match
								FROM users
								WHERE login = $2;`, password, login)

	err = row.Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		ok = false
		return
	}
	return
}

func (db *PostgreSQL) GetUserBalance(login string) (balance add.Balance, err error) {

	row := db.dbConn.QueryRow(`SELECT Sum, With_Drawn FROM Users WHERE login = $1`, login)
	err = row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return
	}

	return
}

func (db *PostgreSQL) GetAllOrders(orders *[]add.Order, login string) (err error) {
	var order add.Order
	rows, err := db.dbConn.Query("SELECT id, status, uploaded_at, accrual FROM orders WHERE user_id=(SELECT id FROM users WHERE login=$1) ORDER BY uploaded_at DESC", login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no orders for user have been found %w", err)
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&order.Number, &order.Status, &order.Uploaded_at, &order.Accrual)
		if err != nil {
			return
		}

		*orders = append(*orders, order)
	}

	return
}

func (db *PostgreSQL) GetSpendOrders(orders *[]add.OrderSpend, login string) (err error) {
	var order add.OrderSpend
	rows, err := db.dbConn.Query("SELECT id, processed_at, sum FROM order_spend WHERE user_id=(SELECT id FROM users WHERE login=$1) ORDER BY orders.uploaded_at DESC", login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no orders for user have been found %w", err)
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&order.Number, &order.Sum, &order.Processed_at)
		if err != nil {
			return
		}

		*orders = append(*orders, order)
	}

	return
}
