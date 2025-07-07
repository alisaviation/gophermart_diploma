package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/database/postgres"
	"github.com/alisaviation/internal/gophermart/dto"
	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/pkg/logger"
)

type OrderService interface {
	UploadOrder(userID int, orderNumber string, goods []dto.AccrualGood) (int, error)
	GetOrders(userID int) ([]models.Order, error)
	StartStatusUpdater(ctx context.Context, interval time.Duration)
}

type OrdersService struct {
	OrderDB       database.Order
	AccrualClient *AccrualClient
}

func NewOrderService(orderDB database.Order, accrualClient *AccrualClient) OrderService {
	return &OrdersService{
		OrderDB:       orderDB,
		AccrualClient: accrualClient,
	}
}

func (s *OrdersService) ValidateOrderNumber(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit / 10) + (digit % 10)
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func (s *OrdersService) UploadOrder(userID int, orderNumber string, goods []dto.AccrualGood) (int, error) {
	if _, err := strconv.Atoi(orderNumber); err != nil {
		return http.StatusBadRequest, errors.New("order number must contain only digits")
	}

	if !s.ValidateOrderNumber(orderNumber) {
		return http.StatusUnprocessableEntity, errors.New("invalid order number by Luhn algorithm")
	}

	existingOrder, err := s.OrderDB.GetOrderByNumber(orderNumber)
	if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return http.StatusInternalServerError, err
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return http.StatusOK, nil
		}
		return http.StatusConflict, errors.New("order number already exists for another user")
	}

	order := &models.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}

	if err := s.OrderDB.CreateOrder(order); err != nil {
		return http.StatusInternalServerError, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.AccrualClient.RegisterOrder(ctx, orderNumber, goods); err != nil {
			s.OrderDB.UpdateOrderStatus(orderNumber, "INVALID")
			return
		}
	}()
	logger.Log.Info("Order accepted for processing",
		zap.String("order", orderNumber),
		zap.Int("user_id", userID))
	return http.StatusAccepted, nil
}

func (s *OrdersService) GetOrders(userID int) ([]models.Order, error) {
	orders, err := s.OrderDB.GetOrdersByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	// Для каждого заказа проверяем актуальную информацию в accrual системе
	for i, order := range orders {
		// Проверяем только заказы, которые ещё не в финальном статусе
		if order.Status == "PROCESSED" || order.Status == "INVALID" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		accrualInfo, err := s.AccrualClient.GetOrderAccrual(ctx, order.Number)
		if err != nil {
			logger.Log.Warn("Failed to get accrual info",
				zap.String("order", order.Number),
				zap.Error(err))
			continue
		}

		if accrualInfo == nil {
			// Заказ не найден в accrual системе
			continue
		}

		// Обновляем статус и accrual, если они изменились
		if order.Status != accrualInfo.Status || order.Accrual != accrualInfo.Accrual {
			if err := s.OrderDB.UpdateOrderFromAccrual(
				order.Number,
				accrualInfo.Status,
				accrualInfo.Accrual,
			); err != nil {
				logger.Log.Error("Failed to update order from accrual",
					zap.String("order", order.Number),
					zap.Error(err))
				continue
			}

			// Обновляем данные в возвращаемом списке
			orders[i].Status = accrualInfo.Status
			orders[i].Accrual = accrualInfo.Accrual
		}
	}
	return orders, nil
}
