package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/logger"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
)

// OrderProcessor обрабатывает заказы в фоновом режиме
type OrderProcessor struct {
	storage        storage.Storage
	accrualService services.AccrualServiceIface
	interval       time.Duration
	stopChan       chan struct{}
	workerCount    int
	rateLimitChan  chan time.Duration
}

// NewOrderProcessor создает новый процессор заказов
func NewOrderProcessor(storage storage.Storage, accrualService services.AccrualServiceIface, interval time.Duration, workerCount int) *OrderProcessor {
	return &OrderProcessor{
		storage:        storage,
		accrualService: accrualService,
		interval:       interval,
		stopChan:       make(chan struct{}),
		workerCount:    workerCount,
		rateLimitChan:  make(chan time.Duration, 1),
	}
}

// Start запускает обработку заказов
func (p *OrderProcessor) Start() {
	go p.processLoop()
}

// Stop останавливает обработку заказов
func (p *OrderProcessor) Stop() {
	close(p.stopChan)
}

// processLoop основной цикл обработки
func (p *OrderProcessor) processLoop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.ProcessOrders()
		case cooldown := <-p.rateLimitChan:
			logger.Logger.Info("Rate limit detected, waiting for cooldown", zap.Duration("cooldown", cooldown))
			time.Sleep(cooldown)
		case <-p.stopChan:
			return
		}
	}
}

// ProcessOrders обрабатывает заказы со статусом NEW и PROCESSING
func (p *OrderProcessor) ProcessOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const batchSize = 100
	offset := 0

	for {
		// Получаем заказы со статусом NEW и PROCESSING с пагинацией
		orders, err := p.storage.GetOrdersByStatusPaginated(ctx, []string{"NEW", "PROCESSING"}, batchSize, offset)
		if err != nil {
			logger.Logger.Error("Failed to get orders for processing", zap.Error(err))
			return
		}

		if len(orders) == 0 {
			break
		}

		logger.Logger.Info("Processing batch of orders",
			zap.Int("count", len(orders)),
			zap.Int("offset", offset),
			zap.Int("workers", p.workerCount))

		p.ProcessOrdersWithWorkers(ctx, orders)

		if len(orders) < batchSize {
			break
		}

		offset += batchSize
	}
}

// ProcessOrdersWithWorkers обрабатывает заказы параллельно
func (p *OrderProcessor) ProcessOrdersWithWorkers(ctx context.Context, orders []models.Order) {
	// Канал для передачи заказов воркерам
	orderChan := make(chan models.Order, len(orders))

	// Канал для сигнализации о rate limit
	rateLimitChan := make(chan struct{}, 1)

	// WaitGroup для ожидания завершения всех воркеров
	var wg sync.WaitGroup

	// Запускаем воркеры
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go p.worker(ctx, orderChan, rateLimitChan, &wg)
	}

	// Отправляем заказы в канал
	go func() {
		defer close(orderChan)
		for _, order := range orders {
			select {
			case orderChan <- order:
			case <-rateLimitChan:
				// Получили сигнал о rate limit, прекращаем отправку
				logger.Logger.Info("Rate limit detected, stopping order processing")
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

// worker обрабатывает заказы из канала
func (p *OrderProcessor) worker(ctx context.Context, orderChan <-chan models.Order, rateLimitChan chan<- struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case order, ok := <-orderChan:
			if !ok {
				return
			}

			if err := p.ProcessOrder(ctx, order.Number); err != nil {
				var rateLimitErr *services.RateLimitError
				if errors.As(err, &rateLimitErr) {
					logger.Logger.Info("Rate limit exceeded, worker stopping",
						zap.Duration("retryAfter", rateLimitErr.RetryAfter))
					// Отправляем cooldown в основной цикл
					select {
					case p.rateLimitChan <- rateLimitErr.RetryAfter:
					default:
					}
					// Отправляем сигнал о rate limit неблокирующе
					select {
					case rateLimitChan <- struct{}{}:
					default:
					}
					return
				}
				logger.Logger.Error("Failed to process order",
					zap.String("orderNumber", order.Number),
					zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

// ProcessOrder обрабатывает конкретный заказ
func (p *OrderProcessor) ProcessOrder(ctx context.Context, orderNumber string) error {
	// Получаем информацию о заказе из системы начисления
	accrualInfo, err := p.accrualService.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		// Если ошибка связана с превышением лимита запросов, не обновляем статус
		if errors.Is(err, services.ErrRateLimitExceeded) {
			return err
		}

		// Обновляем статус на INVALID
		return p.storage.UpdateOrderStatus(ctx, orderNumber, "INVALID", nil)
	}

	if accrualInfo == nil {
		// Заказ не найден в системе начисления
		return p.storage.UpdateOrderStatus(ctx, orderNumber, "INVALID", nil)
	}

	// Обновляем статус и начисление
	var accrual *float64
	if accrualInfo.Accrual != nil {
		accrual = accrualInfo.Accrual
	}

	// Обновляем статус заказа
	if err := p.storage.UpdateOrderStatus(ctx, orderNumber, accrualInfo.Status, accrual); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Если заказ обработан и есть начисление, обновляем баланс пользователя
	if accrualInfo.Status == "PROCESSED" && accrual != nil && *accrual > 0 {
		// Получаем заказ для определения пользователя
		order, err := p.storage.GetOrderByNumber(ctx, orderNumber)
		if err != nil {
			return fmt.Errorf("failed to get order for balance update: %w", err)
		}

		// Получаем текущий баланс
		balance, err := p.storage.GetBalance(ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		// Обновляем баланс
		newCurrent := balance.Current + *accrual
		if err := p.storage.UpdateBalance(ctx, order.UserID, newCurrent, balance.Withdrawn); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}
	}

	return nil
}
