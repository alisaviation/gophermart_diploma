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
														Uploaded_at TIMESTAMP,
														Accrual FLOAT8,
														User_ID BIGINT REFERENCES Users (ID) ON DELETE CASCADE);`)

	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE Order_Spend (ID BIGINT PRIMARY KEY,
													Processed_at TIMESTAMP,
													Sum FLOAT8
													User_ID BIGINT REFERENCES Users (ID) ON DELETE CASCADE);`)
	if err != nil {
		return err
	}

	return nil
}
