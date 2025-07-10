package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
)

// OrderProcessor обрабатывает заказы в фоновом режиме
type OrderProcessor struct {
	storage        storage.Storage
	accrualService services.AccrualServiceIface
	interval       time.Duration
	stopChan       chan struct{}
}

// NewOrderProcessor создает новый процессор заказов
func NewOrderProcessor(storage storage.Storage, accrualService services.AccrualServiceIface, interval time.Duration) *OrderProcessor {
	return &OrderProcessor{
		storage:        storage,
		accrualService: accrualService,
		interval:       interval,
		stopChan:       make(chan struct{}),
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
			p.processOrders()
		case <-p.stopChan:
			return
		}
	}
}

// processOrders обрабатывает заказы со статусом NEW и PROCESSING
func (p *OrderProcessor) processOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Получаем заказы со статусом NEW и PROCESSING
	orders, err := p.storage.GetOrdersByStatus(ctx, []string{"NEW", "PROCESSING"})
	if err != nil {
		log.Printf("Failed to get orders for processing: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	log.Printf("Processing %d orders...", len(orders))

	// Обрабатываем каждый заказ
	for _, order := range orders {
		if err := p.ProcessOrder(ctx, order.Number); err != nil {
			log.Printf("Failed to process order %s: %v", order.Number, err)
			continue
		}
	}
}

// ProcessOrder обрабатывает конкретный заказ
func (p *OrderProcessor) ProcessOrder(ctx context.Context, orderNumber string) error {
	// Получаем информацию о заказе из системы начисления
	accrualInfo, err := p.accrualService.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		// Если ошибка связана с превышением лимита запросов, не обновляем статус
		if err.Error() == "rate limit exceeded" {
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
