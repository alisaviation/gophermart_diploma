package interfaces

import (
	"database/sql"
	"github.com/jackc/pgx/v4"
)

type DBStorageInterface interface {
	Init(connectionString string) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(query string, args ...interface{}) (pgx.Rows, error)
	Close()
	BeginTransaction() error
	Rollback() error
	Commit() error
	QueryRow(query string, args ...interface{}) pgx.Row
}
