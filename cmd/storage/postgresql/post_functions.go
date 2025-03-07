package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (db *PostgreSQL) RegisterNewUser(user add.User) error {

	_, err := db.dbConn.Exec("INSERT INTO users (Login, Password, Sum, With_Drawn) VALUES($1,crypt($2, gen_salt('xdes')),$3,$4)", user.Login, user.Password, 0, 0)

	if err != nil {
		return fmt.Errorf("error while inserting user with login %s: %w", user.Login, err)
	}
	return nil
}

func (db *PostgreSQL) AddNewOrder(login string, orderNumber int) (err error) {

	var id int

	rows, err := db.dbConn.Query("SELECT order_id FROM orders JOIN users ON user_id=users.id WHERE users.Login=$1", login)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return
		}
		if id == orderNumber {
			return fmt.Errorf("the order with number %d is in the system", orderNumber)
		}
	}

	err = rows.Err()
	if err != nil {
		return
	}
	row := db.dbConn.QueryRow("SELECT user_id from orders WHERE order_id=$1", orderNumber)
	err = row.Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	if err == nil {
		return fmt.Errorf("error: order with number %d already exists and belongs to another user", orderNumber)
	}

	_, err = db.dbConn.Exec("INSERT INTO orders (id, status, accrual, uploaded_at, user_id) VALUES($1, $2, $3, $4, (SELECT id FROM users WHERE users.Login=$5))", orderNumber, "NEW", 0, time.Now().Format(time.RFC3339), login)

	return
}

func (db *PostgreSQL) ProcessPayPoints(order add.OrderSpend, login string) (err error) {
	return
}
