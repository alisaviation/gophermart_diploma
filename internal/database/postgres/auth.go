package postgres

import (
	"database/sql"

	"github.com/alisaviation/internal/gophermart/models"
)

func (p *PostgresStorage) CreateUser(user models.User) (int, error) {
	var id int
	err := p.db.QueryRow(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		user.Login, user.PasswordHash,
	).Scan(&id)
	return id, err
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
