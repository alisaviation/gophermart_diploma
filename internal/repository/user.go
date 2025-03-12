package repository

import (
	"github.com/rtmelsov/GopherMart/internal/models"
)

func createDBUser(user *models.User) models.DBUser {
	return models.DBUser{
		Login:    user.Login,
		Password: user.Password,
	}
}

func (r *Repository) Register(user *models.User) (*models.DBUser, *models.Error) {
	body := createDBUser(user)
	return r.db.Register(&body)
}

func (r *Repository) Login(user *models.User) (*models.DBUser, *models.Error) {
	body := createDBUser(user)
	return r.db.Login(&body)
}
