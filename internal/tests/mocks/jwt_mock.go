package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID int, login string) (string, error) {
	args := m.Called(userID, login)
	return args.String(0), args.Error(1)
}
