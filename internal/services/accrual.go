package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// AccrualService сервис для работы с системой начисления баллов
type AccrualService struct {
	client  *http.Client
	baseURL string
	// Retry настройки
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// NewAccrualService создает новый сервис начисления баллов
func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:    baseURL,
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
	}
}

// NewAccrualServiceWithRetry создает сервис с настраиваемыми параметрами retry
func NewAccrualServiceWithRetry(baseURL string, maxRetries int, baseDelay, maxDelay time.Duration) *AccrualService {
	return &AccrualService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:    baseURL,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
	}
}

// shouldRetry определяет, нужно ли повторять запрос
func (s *AccrualService) shouldRetry(err error, statusCode int) bool {
	if err != nil {
		// Повторяем при сетевых ошибках
		return true
	}

	// Повторяем при временных ошибках сервера
	switch statusCode {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}

	return false
}

// calculateDelay вычисляет задержку с jitter для избежания thundering herd
func (s *AccrualService) calculateDelay(attempt int) time.Duration {
	delay := s.baseDelay * time.Duration(1<<attempt) // Экспоненциальная задержка
	if delay > s.maxDelay {
		delay = s.maxDelay
	}

	// Добавляем jitter (±25%)
	jitter := time.Duration(rand.Int63n(int64(delay/2))) - delay/4
	return delay + jitter
}

// GetOrderInfo получает информацию о заказе из системы начисления с retry логикой
func (s *AccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		// Проверяем контекст перед каждым запросом
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Запрос
		url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Выполняем запрос
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			if attempt < s.maxRetries && s.shouldRetry(err, 0) {
				delay := s.calculateDelay(attempt)
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		}
		defer resp.Body.Close()

		// Обрабатываем ответ
		switch resp.StatusCode {
		case http.StatusOK:
			var accrualResp models.AccrualResponse
			if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}
			return &accrualResp, nil
		case http.StatusNoContent:
			return nil, nil
		case http.StatusTooManyRequests:
			// Получаем Retry-After заголовок
			retryAfterStr := resp.Header.Get("Retry-After")
			var retryAfter time.Duration
			if retryAfterStr != "" {
				if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
					retryAfter = time.Duration(seconds) * time.Second
				}
			}
			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError:
			lastErr = ErrInternalServer
			if attempt < s.maxRetries && s.shouldRetry(nil, resp.StatusCode) {
				delay := s.calculateDelay(attempt)
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			if attempt < s.maxRetries && s.shouldRetry(nil, resp.StatusCode) {
				delay := s.calculateDelay(attempt)
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		default:
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return nil, lastErr
}
