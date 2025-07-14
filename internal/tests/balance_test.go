package tests

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/tests/mocks"
)

func TestBalancesService_CreateWithdrawal(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mocks.MockBalance)
		args        *models.Withdrawal
		wantErr     bool
		errContains string
	}{
		{
			name: "successful withdrawal creation",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(false, nil)
				mb.On("CreateWithdrawal", &models.Withdrawal{
					UserID:      1,
					OrderNumber: "123",
					Sum:         100.5,
				}).Return(nil)
			},
			args: &models.Withdrawal{
				UserID:      1,
				OrderNumber: "123",
				Sum:         100.5,
			},
			wantErr: false,
		},
		{
			name: "withdrawal already exists",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(true, nil)
			},
			args: &models.Withdrawal{
				OrderNumber: "123",
			},
			wantErr:     true,
			errContains: "withdrawal for order 123 already exists",
		},
		{
			name: "error checking withdrawal existence",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(false, errors.New("database error"))
			},
			args: &models.Withdrawal{
				OrderNumber: "123",
			},
			wantErr:     true,
			errContains: "failed to check withdrawal existence",
		},
		{
			name: "error creating withdrawal",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(false, nil)
				mb.On("CreateWithdrawal", mock.Anything).Return(errors.New("database error"))
			},
			args: &models.Withdrawal{
				OrderNumber: "123",
			},
			wantErr:     true,
			errContains: "failed to create withdrawal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBalance := new(mocks.MockBalance)
			tt.setupMock(mockBalance)

			s := &services.BalancesService{
				Balance: mockBalance,
			}

			err := s.CreateWithdrawal(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWithdrawal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CreateWithdrawal() error = %v, should contain %v", err, tt.errContains)
				}
			}

			mockBalance.AssertExpectations(t)
		})
	}
}

func TestBalancesService_GetUserBalance(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockBalance)
		userID    int
		want      *models.Balance
		wantErr   bool
	}{
		{
			name: "successful balance retrieval",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("GetBalance", 1).Return(&models.Balance{
					Current:   500.75,
					Withdrawn: 100.25,
				}, nil)
			},
			userID: 1,
			want: &models.Balance{
				Current:   500.75,
				Withdrawn: 100.25,
			},
			wantErr: false,
		},
		{
			name: "error getting balance",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("GetBalance", 1).Return(nil, errors.New("database error"))
			},
			userID:  1,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBalance := new(mocks.MockBalance)
			tt.setupMock(mockBalance)

			s := &services.BalancesService{
				Balance: mockBalance,
			}

			got, err := s.GetUserBalance(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserBalance() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserBalance() got = %v, want %v", got, tt.want)
			}

			mockBalance.AssertExpectations(t)
		})
	}
}

func TestBalancesService_GetUserWithdrawals(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockBalance)
		userID    int
		want      []models.Withdrawal
		wantErr   bool
	}{
		{
			name: "successful withdrawals retrieval",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("GetWithdrawals", 1).Return([]models.Withdrawal{
					{
						UserID:      1,
						OrderNumber: "123",
						Sum:         100.5,
					},
					{
						UserID:      1,
						OrderNumber: "456",
						Sum:         200.75,
					},
				}, nil)
			},
			userID: 1,
			want: []models.Withdrawal{
				{
					UserID:      1,
					OrderNumber: "123",
					Sum:         100.5,
				},
				{
					UserID:      1,
					OrderNumber: "456",
					Sum:         200.75,
				},
			},
			wantErr: false,
		},
		{
			name: "no withdrawals found",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("GetWithdrawals", 1).Return([]models.Withdrawal{}, nil)
			},
			userID:  1,
			want:    []models.Withdrawal{},
			wantErr: false,
		},
		{
			name: "error getting withdrawals",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("GetWithdrawals", 1).Return(nil, errors.New("database error"))
			},
			userID:  1,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBalance := new(mocks.MockBalance)
			tt.setupMock(mockBalance)

			s := &services.BalancesService{
				Balance: mockBalance,
			}

			got, err := s.GetUserWithdrawals(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserWithdrawals() got = %v, want %v", got, tt.want)
			}

			mockBalance.AssertExpectations(t)
		})
	}
}

func TestBalancesService_WithdrawalExists(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mocks.MockBalance)
		orderNumber string
		want        bool
		wantErr     bool
	}{
		{
			name: "withdrawal exists",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(true, nil)
			},
			orderNumber: "123",
			want:        true,
			wantErr:     false,
		},
		{
			name: "withdrawal does not exist",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(false, nil)
			},
			orderNumber: "123",
			want:        false,
			wantErr:     false,
		},
		{
			name: "error checking withdrawal",
			setupMock: func(mb *mocks.MockBalance) {
				mb.On("WithdrawalExists", "123").Return(false, errors.New("database error"))
			},
			orderNumber: "123",
			want:        false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBalance := new(mocks.MockBalance)
			tt.setupMock(mockBalance)

			s := &services.BalancesService{
				Balance: mockBalance,
			}

			got, err := s.WithdrawalExists(tt.orderNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawalExists() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("WithdrawalExists() got = %v, want %v", got, tt.want)
			}

			mockBalance.AssertExpectations(t)
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
