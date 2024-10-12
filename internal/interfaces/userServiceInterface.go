package interfaces

import (
	"gophermart/internal/models"
)

type UserServiceInterface interface {
	GetUserID(username string) int
	RegisterUser(user models.User) (models.User, error)
	AuthenticateUser(username, password string) (models.User, error)
	GetUserRepository() UserRepositoryInterface
}
