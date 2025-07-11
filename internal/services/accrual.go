package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// AccrualService сервис для работы с системой начисления баллов
type AccrualService struct {
	client  *http.Client
	baseURL string
}

// NewAccrualService создает новый сервис начисления баллов
func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetOrderInfo получает информацию о заказе из системы начисления
func (s *AccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

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
		return nil, ErrInternalServer
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
