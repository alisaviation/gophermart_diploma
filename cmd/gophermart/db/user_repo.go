package db

import (
	"database/sql"
)

type UserRepoPG struct {
	db *sql.DB
}

func NewUserRepoPG(db *sql.DB) *UserRepoPG {
	return &UserRepoPG{db: db}
}

func (r *UserRepoPG) CreateUser(login, passwordHash string) error {
	_, err := r.db.Exec(`INSERT INTO users (login, password_hash) VALUES ($1, $2)`, login, passwordHash)
	return err
}

func (r *UserRepoPG) IsLoginExist(login string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)`, login).Scan(&exists)
	return exists, err
}
