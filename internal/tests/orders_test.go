package tests

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/database/postgres"
	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/internal/gophermart/services"
)

func Test_orderService_GetOrders(t *testing.T) {
	mockOrderDB := new(MockOrderDB)
	now := time.Now()
	testOrders := []models.Order{
		{
			UserID:     1,
			Number:     "1234567890",
			Status:     "PROCESSED",
			UploadedAt: now.Add(-24 * time.Hour),
		},
		{
			UserID:     1,
			Number:     "9876543210",
			Status:     "NEW",
			UploadedAt: now,
		},
	}

	type fields struct {
		orderDB database.Order
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        []models.Order
		wantErr     bool
		expectedErr string
		setupMock   func()
	}{
		{
			name: "successful get orders",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID: 1,
			},
			want:    testOrders,
			wantErr: false,
			setupMock: func() {
				mockOrderDB.On("GetOrdersByUser", 1).Return(testOrders, nil)
			},
		},
		{
			name: "empty orders list",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID: 2,
			},
			want:    []models.Order{},
			wantErr: false,
			setupMock: func() {
				mockOrderDB.On("GetOrdersByUser", 2).Return([]models.Order{}, nil)
			},
		},
		{
			name: "database error",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID: 3,
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "failed to get user orders: database error",
			setupMock: func() {
				mockOrderDB.On("GetOrdersByUser", 3).Return(nil, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderDB.ExpectedCalls = nil
			tt.setupMock()

			s := &services.OrdersService{
				OrderDB: tt.fields.orderDB,
			}
			got, err := s.GetOrders(tt.args.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("GetOrders() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrders() got = %v, want %v", got, tt.want)
			}
			mockOrderDB.AssertExpectations(t)
		})
	}
}

func Test_orderService_UploadOrder(t *testing.T) {
	mockOrderDB := new(MockOrderDB)

	type fields struct {
		orderDB database.Order
	}
	type args struct {
		userID      int
		orderNumber string
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        int
		wantErr     bool
		expectedErr error
		setupMock   func()
	}{
		{
			name: "successful order upload",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "4561261212345467",
			},
			want:    http.StatusAccepted,
			wantErr: false,
			setupMock: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").Return(nil, postgres.ErrNotFound)
				mockOrderDB.On("CreateOrder", mock.Anything).Return(nil)
			},
		},
		{
			name: "invalid order number - non digits",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "123abc456",
			},
			want:        http.StatusBadRequest,
			wantErr:     true,
			expectedErr: errors.New("order number must contain only digits"),
			setupMock:   func() {},
		},
		{
			name: "invalid order number - fails Luhn check",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "4561261212345464",
			},
			want:        http.StatusUnprocessableEntity,
			wantErr:     true,
			expectedErr: errors.New("invalid order number by Luhn algorithm"),
			setupMock:   func() {},
		},
		{
			name: "order already exists for same user",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "79927398713",
			},
			want:    http.StatusOK,
			wantErr: false,
			setupMock: func() {
				mockOrderDB.On("GetOrderByNumber", "79927398713").Return(&models.Order{
					UserID: 1,
					Number: "79927398713",
				}, nil)
			},
		},
		{
			name: "order already exists for different user",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      2,
				orderNumber: "79927398713",
			},
			want:        http.StatusConflict,
			wantErr:     true,
			expectedErr: errors.New("order number already exists for another user"),
			setupMock: func() {
				mockOrderDB.On("GetOrderByNumber", "79927398713").Return(&models.Order{
					UserID: 1,
					Number: "79927398713",
				}, nil)
			},
		},
		{
			name: "database error when getting order",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "4561261212345467",
			},
			want:        http.StatusInternalServerError,
			wantErr:     true,
			expectedErr: errors.New("database error"),
			setupMock: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").Return(nil, errors.New("database error"))
			},
		},
		{
			name: "database error when creating order",
			fields: fields{
				orderDB: mockOrderDB,
			},
			args: args{
				userID:      1,
				orderNumber: "4561261212345467",
			},
			want:        http.StatusInternalServerError,
			wantErr:     true,
			expectedErr: errors.New("database error"),
			setupMock: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").Return(nil, postgres.ErrNotFound)
				mockOrderDB.On("CreateOrder", mock.Anything).Return(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderDB.ExpectedCalls = nil
			tt.setupMock()

			s := &services.OrdersService{
				OrderDB: tt.fields.orderDB,
			}
			got, err := s.UploadOrder(tt.args.userID, tt.args.orderNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("UploadOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UploadOrder() got = %v, want %v", got, tt.want)
			}
			if tt.wantErr && err.Error() != tt.expectedErr.Error() {
				t.Errorf("UploadOrder() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			mockOrderDB.AssertExpectations(t)
		})
	}
}

func Test_orderService_ValidateOrderNumber(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid order number",
			args: args{number: "4561261212345467"},
			want: true,
		},
		{
			name: "invalid order number",
			args: args{number: "4561261212345464"},
			want: false,
		},
		{
			name: "valid but with spaces",
			args: args{number: " 4561261212345467 "},
			want: false,
		},
		{
			name: "empty string",
			args: args{number: ""},
			want: false,
		},
		{
			name: "non-numeric characters",
			args: args{number: "4561a61212345467"},
			want: false,
		},
		{
			name: "another valid number",
			args: args{number: "79927398713"},
			want: true,
		},
		{
			name: "short invalid number",
			args: args{number: "42"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &services.OrdersService{}
			if got := s.ValidateOrderNumber(tt.args.number); got != tt.want {
				t.Errorf("ValidateOrderNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
