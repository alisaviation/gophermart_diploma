package tests

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/internal/database/postgres"
	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/tests/mocks"
)

func TestOrderService_UploadOrder(t *testing.T) {
	mockOrderDB := new(mocks.MockOrderDB)
	mockAccrualClient := new(mocks.MockAccrualClient)

	orderService := services.NewOrderService(mockOrderDB, mockAccrualClient)

	tests := []struct {
		name           string
		userID         int
		orderNumber    string
		mockSetup      func()
		expectedStatus int
		expectedError  error
	}{
		{
			name:           "invalid order number - non digits",
			userID:         1,
			orderNumber:    "123abc",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  errors.New("order number must contain only digits"),
		},
		{
			name:           "invalid order number - fails Luhn check",
			userID:         1,
			orderNumber:    "1234567890",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  errors.New("invalid order number by Luhn algorithm"),
		},
		{
			name:        "order exists for same user",
			userID:      1,
			orderNumber: "4561261212345467",
			mockSetup: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").
					Return(&models.Order{UserID: 1, Number: "4561261212345467"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:        "order exists for different user",
			userID:      1,
			orderNumber: "4561261212345467",
			mockSetup: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").
					Return(&models.Order{UserID: 2, Number: "4561261212345467"}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  errors.New("order number already exists for another user"),
		},
		{
			name:        "successful order upload",
			userID:      1,
			orderNumber: "4561261212345467",
			mockSetup: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").
					Return(nil, postgres.ErrNotFound)
				mockOrderDB.On("CreateOrder", mock.AnythingOfType("*models.Order")).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedError:  nil,
		},
		{
			name:        "database error when checking order",
			userID:      1,
			orderNumber: "4561261212345467",
			mockSetup: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  errors.New("failed to check order: database error"),
		},
		{
			name:        "database error when creating order",
			userID:      1,
			orderNumber: "4561261212345467",
			mockSetup: func() {
				mockOrderDB.On("GetOrderByNumber", "4561261212345467").
					Return(nil, postgres.ErrNotFound)
				mockOrderDB.On("CreateOrder", mock.AnythingOfType("*models.Order")).
					Return(errors.New("failed to create order: database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  errors.New("failed to create order: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderDB.ExpectedCalls = nil
			mockAccrualClient.ExpectedCalls = nil

			tt.mockSetup()

			status, err := orderService.UploadOrder(tt.userID, tt.orderNumber)

			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockOrderDB.AssertExpectations(t)
			mockAccrualClient.AssertExpectations(t)
		})
	}
}

func TestOrderService_GetOrders(t *testing.T) {
	mockOrderDB := new(mocks.MockOrderDB)
	mockAccrualClient := new(mocks.MockAccrualClient)

	orderService := services.NewOrderService(mockOrderDB, mockAccrualClient)

	now := time.Now()

	tests := []struct {
		name          string
		userID        int
		mockSetup     func()
		expected      []models.Order
		expectedError error
	}{
		{
			name:   "successful get orders",
			userID: 1,
			mockSetup: func() {
				mockOrderDB.On("GetOrdersByUser", 1).
					Return([]models.Order{
						{
							UserID:     1,
							Number:     "123",
							Status:     "NEW",
							Accrual:    0,
							UploadedAt: now,
						},
						{
							UserID:     1,
							Number:     "456",
							Status:     "PROCESSED",
							Accrual:    100,
							UploadedAt: now.Add(-time.Hour),
						},
					}, nil)

				mockAccrualClient.On("GetOrderAccrual", mock.Anything, "123").
					Return(&dto.AccrualResponse{
						Order:   "123",
						Status:  "NEW",
						Accrual: 0,
					}, nil)
			},
			expected: []models.Order{
				{
					UserID:     1,
					Number:     "123",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: now,
				},
				{
					UserID:     1,
					Number:     "456",
					Status:     "PROCESSED",
					Accrual:    100,
					UploadedAt: now.Add(-time.Hour),
				},
			},
			expectedError: nil,
		},
		{
			name:   "database error",
			userID: 1,
			mockSetup: func() {
				mockOrderDB.On("GetOrdersByUser", 1).
					Return(nil, errors.New("database error"))
			},
			expected:      nil,
			expectedError: errors.New("failed to get user orders: database error"),
		},
		{
			name:   "update order status from accrual",
			userID: 1,
			mockSetup: func() {
				mockOrderDB.On("GetOrdersByUser", 1).
					Return([]models.Order{
						{
							UserID:     1,
							Number:     "123",
							Status:     "NEW",
							Accrual:    0,
							UploadedAt: now,
						},
					}, nil)

				mockAccrualClient.On("GetOrderAccrual", mock.Anything, "123").
					Return(&dto.AccrualResponse{
						Order:   "123",
						Status:  "PROCESSED",
						Accrual: 50,
					}, nil)

				mockOrderDB.On("UpdateOrderFromAccrual", "123", "PROCESSED", 50.0).
					Return(nil)
			},
			expected: []models.Order{
				{
					UserID:     1,
					Number:     "123",
					Status:     "PROCESSED",
					Accrual:    50,
					UploadedAt: now,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderDB.ExpectedCalls = nil
			mockAccrualClient.ExpectedCalls = nil

			tt.mockSetup()

			orders, err := orderService.GetOrders(tt.userID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, orders)
			}

			mockOrderDB.AssertExpectations(t)
			mockAccrualClient.AssertExpectations(t)
		})
	}
}

func TestOrderService_ValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid order number",
			number: "4561261212345467",
			want:   true,
		},
		{
			name:   "invalid order number",
			number: "4561261212345464",
			want:   false,
		},
		{
			name:   "valid but with spaces",
			number: " 4561261212345467 ",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "non-numeric characters",
			number: "4561a61212345467",
			want:   false,
		},
		{
			name:   "another valid number",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "short number",
			number: "42",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.ValidateOrderNumber(tt.number)
			assert.Equal(t, tt.want, got)
		})
	}
}
