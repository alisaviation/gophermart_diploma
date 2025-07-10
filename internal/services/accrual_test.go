package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccrualService(t *testing.T) {
	baseURL := "http://localhost:8080"
	service := NewAccrualService(baseURL)

	assert.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
	assert.NotNil(t, service.client)
}

func TestAccrualService_GetOrderInfo_Success(t *testing.T) {
	// Создаем тестовый сервер
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
