package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQL struct {
	Address  string
	Port     string
	UserName string
	Password string
	DBName   string
	dbConn   *sql.DB
}

func (db *PostgreSQL) Init() error {
	var err error
	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable",
		db.Address, db.Port, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}

	// _, err = db.dbConn.Exec(`CREATE EXTENSION pgcrypto;`)
	// if err != nil {
	// 	return fmt.Errorf("error while creating extension pgcrypto: %w", err)
	// }

	_, err = db.dbConn.Exec(`CREATE TABLE Users (ID BIGSERIAL PRIMARY KEY,
												Login VARCHAR(1000) NOT NULL UNIQUE,
												Password VARCHAR(1000) NOT NULL,
	                                            Sum FLOAT8, 
												With_Drawn FLOAT8);`)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TYPE Status_Enum AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');`)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE Orders (ID BIGINT PRIMARY KEY,
														Status Status_Enum,
														Uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
														Accrual BIGINT);`)

	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE Users_To_Orders (User_ID BIGINT REFERENCES Users (ID),
														Order_ID BIGINT REFERENCES Orders (ID));`)

	if err != nil {
		return err
	}

	return nil
}
