package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/pkg/logger"
)

type AccrualClient struct {
	baseURL    string
	client     *http.Client
	retryDelay time.Duration
	maxRetries int
}

func NewAccrualClient(baseURL string) *AccrualClient {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &AccrualClient{
		baseURL:    baseURL,
		client:     &http.Client{Timeout: 10 * time.Second},
		retryDelay: 1 * time.Second,
		maxRetries: 3,
	}
}

func (c *AccrualClient) RegisterOrder(ctx context.Context, orderNumber string, goods []dto.AccrualGood) error {
	request := dto.AccrualOrderRequest{
		Order: orderNumber,
		Goods: goods,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/orders", c.baseURL)

	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusAccepted:
			return nil
		case http.StatusBadRequest:
			return fmt.Errorf("invalid request format")
		case http.StatusConflict:
			return fmt.Errorf("order already registered")
		case http.StatusInternalServerError:
			lastErr = fmt.Errorf("accrual server error")
			continue
		default:
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}
	}

	return lastErr
}

func (c *AccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*dto.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Log.Info("Retrying accrual info request",
				zap.String("order", orderNumber),
				zap.Int("attempt", attempt))
			time.Sleep(c.retryDelay)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			logger.Log.Error("Failed to create accrual info request",
				zap.String("order", orderNumber),
				zap.Error(err))
			continue
		}

		startTime := time.Now()
		resp, err := c.client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			logger.Log.Error("Accrual info request failed",
				zap.String("order", orderNumber),
				zap.Duration("duration", duration),
				zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		logger.Log.Info("Accrual info response",
			zap.String("order", orderNumber),
			zap.Int("status_code", resp.StatusCode),
			zap.Duration("duration", duration))

		switch resp.StatusCode {
		case http.StatusOK:
			var accrualResp dto.AccrualResponse
			if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
				lastErr = fmt.Errorf("failed to decode response: %w", err)
				logger.Log.Error("Failed to decode accrual response",
					zap.String("order", orderNumber),
					zap.Error(err))
				continue
			}

			logger.Log.Debug("Successfully got accrual info",
				zap.String("order", orderNumber),
				zap.String("status", accrualResp.Status),
				zap.Float64("accrual", accrualResp.Accrual))

			return &accrualResp, nil

		case http.StatusNoContent:
			logger.Log.Debug("Order not registered in accrual system",
				zap.String("order", orderNumber))
			return nil, nil

		case http.StatusTooManyRequests:
			retryAfter := resp.Header.Get("Retry-After")
			logger.Log.Warn("Rate limit exceeded for accrual info",
				zap.String("order", orderNumber),
				zap.String("retry_after", retryAfter))
			lastErr = fmt.Errorf("rate limit exceeded")
			continue

		case http.StatusInternalServerError:
			logger.Log.Warn("Accrual system internal error",
				zap.String("order", orderNumber))
			lastErr = fmt.Errorf("accrual server error")
			continue

		default:
			logger.Log.Warn("Unexpected status from accrual system",
				zap.String("order", orderNumber),
				zap.Int("status_code", resp.StatusCode))
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}
	}

	logger.Log.Error("All attempts to get accrual info failed",
		zap.String("order", orderNumber),
		zap.Error(lastErr))
	return nil, lastErr
}
