package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAccrualService(t *testing.T) {
	baseURL := "http://localhost:8080"
	service := NewAccrualService(baseURL)

	assert.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
	assert.NotNil(t, service.client)
	assert.Equal(t, 3, service.maxRetries)
	assert.Equal(t, 100*time.Millisecond, service.baseDelay)
	assert.Equal(t, 5*time.Second, service.maxDelay)
}

func TestNewAccrualServiceWithRetry(t *testing.T) {
	baseURL := "http://localhost:8080"
	maxRetries := 5
	baseDelay := 200 * time.Millisecond
	maxDelay := 10 * time.Second

	service := NewAccrualServiceWithRetry(baseURL, maxRetries, baseDelay, maxDelay)

	assert.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
	assert.NotNil(t, service.client)
	assert.Equal(t, maxRetries, service.maxRetries)
	assert.Equal(t, baseDelay, service.baseDelay)
	assert.Equal(t, maxDelay, service.maxDelay)
}

func TestAccrualService_GetOrderInfo_Success(t *testing.T) {
	// Тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"order": "12345678903",
			"status": "PROCESSED",
			"accrual": 500
		}`))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "12345678903", result.Order)
	assert.Equal(t, "PROCESSED", result.Status)
	assert.NotNil(t, result.Accrual)
	assert.Equal(t, 500.0, *result.Accrual)
}

func TestAccrualService_GetOrderInfo_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("No more than N requests per minute allowed"))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Equal(t, "rate limit exceeded", err.Error())
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Equal(t, "internal server error from accrual system", err.Error())
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_NetworkError(t *testing.T) {
	// Используем несуществующий URL для симуляции сетевой ошибки
	service := NewAccrualService("http://localhost:99999")
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Первые два запроса возвращают ошибку
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Третий запрос успешен
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"order": "12345678903",
				"status": "PROCESSED",
				"accrual": 500
			}`))
		}
	}))
	defer server.Close()

	// Создаем сервис с быстрыми retry
	service := NewAccrualServiceWithRetry(server.URL, 3, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "12345678903", result.Order)
	assert.Equal(t, "PROCESSED", result.Status)
	assert.Equal(t, 3, attempts) // Проверяем, что было 3 попытки
}

func TestAccrualService_GetOrderInfo_RetryExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Всегда возвращаем ошибку
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Создаем сервис с быстрыми retry
	service := NewAccrualServiceWithRetry(server.URL, 2, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "internal server error from accrual system", err.Error())
	assert.Equal(t, 3, attempts) // Проверяем, что было 3 попытки (включая первую)
}

func TestAccrualService_GetOrderInfo_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Долгая задержка для симуляции медленного ответа
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
	// Проверяем, что ошибка связана с контекстом
	assert.Contains(t, err.Error(), "context")
}

func TestAccrualService_shouldRetry(t *testing.T) {
	service := NewAccrualService("http://localhost:8080")

	// Сетевые ошибки
	assert.True(t, service.shouldRetry(fmt.Errorf("connection refused"), 0))

	// Временные ошибки сервера
	assert.True(t, service.shouldRetry(nil, http.StatusTooManyRequests))
	assert.True(t, service.shouldRetry(nil, http.StatusInternalServerError))
	assert.True(t, service.shouldRetry(nil, http.StatusBadGateway))
	assert.True(t, service.shouldRetry(nil, http.StatusServiceUnavailable))
	assert.True(t, service.shouldRetry(nil, http.StatusGatewayTimeout))

	// Ошибки, которые не должны повторяться
	assert.False(t, service.shouldRetry(nil, http.StatusOK))
	assert.False(t, service.shouldRetry(nil, http.StatusNotFound))
	assert.False(t, service.shouldRetry(nil, http.StatusBadRequest))
}

func TestAccrualService_calculateDelay(t *testing.T) {
	service := NewAccrualService("http://localhost:8080")

	// Экспоненциальную задержку
	delay1 := service.calculateDelay(0)
	delay2 := service.calculateDelay(1)
	delay3 := service.calculateDelay(2)

	assert.True(t, delay2 > delay1)
	assert.True(t, delay3 > delay2)

	// Задержка не превышает максимальную
	maxDelay := service.calculateDelay(10)
	assert.True(t, maxDelay <= service.maxDelay)
}
