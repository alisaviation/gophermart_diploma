package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// MockStorage мок для Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	args := m.Called(ctx, login, passwordHash)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorage) CreateOrder(ctx context.Context, userID int64, number string) (*models.Order, error) {
	args := m.Called(ctx, userID, number)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockStorage) GetOrdersByStatus(ctx context.Context, statuses []string) ([]models.Order, error) {
	args := m.Called(ctx, statuses)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockStorage) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

func (m *MockStorage) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Balance), args.Error(1)
}

func (m *MockStorage) UpdateBalance(ctx context.Context, userID int64, current, withdrawn float64) error {
	args := m.Called(ctx, userID, current, withdrawn)
	return args.Error(0)
}

func (m *MockStorage) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	args := m.Called(ctx, userID, order, sum)
	return args.Get(0).(*models.Withdrawal), args.Error(1)
}

func (m *MockStorage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

func (m *MockStorage) ProcessWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	args := m.Called(ctx, userID, order, sum)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Withdrawal), args.Error(1)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockAccrualService мок для AccrualServiceIface
type MockAccrualService struct {
	mock.Mock
}

func (m *MockAccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccrualResponse), args.Error(1)
}

func TestNewOrderProcessor(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	interval := 5 * time.Second

	processor := NewOrderProcessor(mockStorage, mockAccrualService, interval)

	assert.NotNil(t, processor)
	assert.Equal(t, mockStorage, processor.storage)
	assert.Equal(t, mockAccrualService, processor.accrualService)
	assert.Equal(t, interval, processor.interval)
	assert.NotNil(t, processor.stopChan)
}

func TestOrderProcessor_StartStop(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	interval := 100 * time.Millisecond

	processor := NewOrderProcessor(mockStorage, mockAccrualService, interval)

	// Запускаем процессор
	processor.Start()

	// Ждем немного, чтобы убедиться, что он запустился
	time.Sleep(50 * time.Millisecond)

	// Останавливаем процессор
	processor.Stop()

	// Ждем немного, чтобы убедиться, что он остановился
	time.Sleep(50 * time.Millisecond)
}

func TestOrderProcessor_ProcessOrder_Success(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second)

	ctx := context.Background()
	orderNumber := "12345678903"
	accrualValue := 100.0
	userID := int64(1)

	// Настраиваем моки
	mockAccrualService.On("GetOrderInfo", ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: &accrualValue,
	}, nil)

	mockStorage.On("UpdateOrderStatus", ctx, orderNumber, "PROCESSED", &accrualValue).Return(nil)

	// Добавляем ожидания для обновления баланса
	mockStorage.On("GetOrderByNumber", ctx, orderNumber).Return(&models.Order{
		ID:     1,
		UserID: userID,
		Number: orderNumber,
		Status: "NEW",
	}, nil)

	mockStorage.On("GetBalance", ctx, userID).Return(&models.Balance{
		UserID:    userID,
		Current:   50.0,
		Withdrawn: 0.0,
	}, nil)

	mockStorage.On("UpdateBalance", ctx, userID, 150.0, 0.0).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_InvalidOrder(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second)

	ctx := context.Background()
	orderNumber := "12345678903"

	// Настраиваем моки - заказ не найден
	mockAccrualService.On("GetOrderInfo", ctx, orderNumber).Return(nil, nil)
	mockStorage.On("UpdateOrderStatus", ctx, orderNumber, "INVALID", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_RateLimitError(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second)

	ctx := context.Background()
	orderNumber := "12345678903"

	// Настраиваем моки - превышение лимита запросов
	mockAccrualService.On("GetOrderInfo", ctx, orderNumber).Return(nil, assert.AnError)
	mockStorage.On("UpdateOrderStatus", ctx, orderNumber, "INVALID", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_ActualRateLimitError(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second)

	ctx := context.Background()
	orderNumber := "12345678903"

	// Настраиваем моки - реальная rate limit ошибка
	mockAccrualService.On("GetOrderInfo", ctx, orderNumber).Return(nil, fmt.Errorf("rate limit exceeded"))

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, "rate limit exceeded", err.Error())
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "UpdateOrderStatus")
}

func TestOrderProcessor_ProcessOrder_NoAccrual(t *testing.T) {
	mockStorage := &MockStorage{}
	mockAccrualService := &MockAccrualService{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second)

	ctx := context.Background()
	orderNumber := "12345678903"

	// Настраиваем моки - заказ обработан, но без начисления
	mockAccrualService.On("GetOrderInfo", ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)

	mockStorage.On("UpdateOrderStatus", ctx, orderNumber, "PROCESSED", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
