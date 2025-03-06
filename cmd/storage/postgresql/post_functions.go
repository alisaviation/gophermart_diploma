package storage

import (
	"fmt"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (db *PostgreSQL) RegisterNewUser(user add.User) error {

	_, err := db.dbConn.Exec("INSERT INTO users (Login, Password, Sum, With_Drawn) VALUES($1,crypt($2, gen_salt('xdes')),$3,$4)", user.Login, user.Password, 0, 0)

	if err != nil {
		return fmt.Errorf("error while inserting user with login %s: %w", user.Login, err)
	}
	return nil
}

func (db *PostgreSQL) AddNewOrder(user add.User, orderNumber int) error {
	// вытащить по ID заказа всех пользователей

	//если строк нет - добавить новый заказ,

	// если строки есть, проверить, что пользователь из БД по ID совпадает с пользователем извне

	// если да - вернуть код 200

	// если нет - вернуть код 409


	return nil
}
