package storage

import (
	"database/sql"
)

type (
	PostgresStorage struct {
		db *sql.DB
	}
)
