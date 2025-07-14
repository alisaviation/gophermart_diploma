package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/dto"
)

type MockAccrualClient struct {
	mock.Mock
}

func (m *MockAccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*dto.AccrualResponse, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AccrualResponse), args.Error(1)
}
