package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/gophermart/models"
	"github.com/alisaviation/pkg/logger"
)

func (s *OrdersService) StartStatusUpdater(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopping status updater")
			return
		case <-ticker.C:
			s.updatePendingOrdersStatus(ctx)
		}
	}
}

func (s *OrdersService) updatePendingOrdersStatus(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	orders, err := s.OrderDB.GetOrdersByStatuses([]string{"NEW"})
	if err != nil {
		logger.Log.Error("Failed to get pending orders for update",
			zap.Error(err))
		return
	}

	logger.Log.Debug("Updating statuses for pending orders",
		zap.Int("count", len(orders)))

	for _, order := range orders {
		select {
		case <-ctx.Done():
			logger.Log.Warn("Context done during status update")
			return
		default:
			s.updateSingleOrderStatus(ctx, order)
		}
	}
}

func (s *OrdersService) updateSingleOrderStatus(ctx context.Context, order models.Order) {

	accrualInfo, err := s.AccrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		logger.Log.Warn("Failed to get accrual info",
			zap.String("order", order.Number),
			zap.Error(err))
		return
	}

	if accrualInfo == nil {
		logger.Log.Debug("No accrual info for order",
			zap.String("order", order.Number))
		return
	}

	if order.Status != accrualInfo.Status || order.Accrual != accrualInfo.Accrual {
		if err := s.OrderDB.UpdateOrderFromAccrual(
			order.Number,
			accrualInfo.Status,
			accrualInfo.Accrual,
		); err != nil {
			logger.Log.Error("Failed to update order",
				zap.String("order", order.Number),
				zap.Error(err))
		} else {
			logger.Log.Info("Order updated",
				zap.String("order", order.Number),
				zap.String("status", accrualInfo.Status),
				zap.Float64("accrual", accrualInfo.Accrual))
		}
	}
}
