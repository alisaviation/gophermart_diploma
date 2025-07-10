package postgres

import (
	"database/sql"

	"github.com/alisaviation/internal/gophermart/models"
)

func (p *PostgresStorage) CreateUser(user models.User) error {
	_, err := p.db.Exec(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		user.Login, user.PasswordHash,
	)
	return err
}

func (p *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow(
		"SELECT id, login, password_hash FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
