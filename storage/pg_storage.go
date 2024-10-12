package storage

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type PgStorage struct {
	Conn        *pgxpool.Pool
	Ctx         context.Context
	transaction pgx.Tx
}

func (pgs *PgStorage) Init(connectionString string) error {
	pgs.Ctx = context.Background()
	var err error
	pgs.Conn, err = pgxpool.Connect(pgs.Ctx, connectionString)

	if err != nil {
		return err
	}

	return nil
}

func (pgs *PgStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (pgs *PgStorage) Select(query string, args ...interface{}) (pgx.Rows, error) {
	rows, err := pgs.Conn.Query(pgs.Ctx, query, args...)
	if err != nil {
		log.Printf("Query execution error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

func (pgs *PgStorage) Close() {
	pgs.Conn.Close()
}

func (pgs *PgStorage) BeginTransaction() error {
	var err error
	pgs.transaction, err = pgs.Conn.BeginTx(
		pgs.Ctx, pgx.TxOptions{})

	return err
}

func (pgs *PgStorage) Rollback() error {
	return pgs.transaction.Rollback(pgs.Ctx)
}

func (pgs *PgStorage) Commit() error {
	return pgs.transaction.Commit(pgs.Ctx)
}

func (pgs *PgStorage) QueryRow(query string, args ...interface{}) pgx.Row {
	return pgs.transaction.QueryRow(pgs.Ctx, query, args...)
}
