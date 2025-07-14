package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/models"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByLogin(login string) (*models.User, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user models.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}
