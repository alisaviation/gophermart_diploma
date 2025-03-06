package storage

import (
	"database/sql"
	"errors"
	"fmt"

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

func (db *PostgreSQL) CheckUser(user add.User) (ok bool, err error) {
	ok = true
	row := db.dbConn.QueryRow(`SELECT (password = crypt($1, password)) 
								AS password_match
								FROM users
								WHERE login = $2;`, user.Password, user.Login)

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

	
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	return
}
