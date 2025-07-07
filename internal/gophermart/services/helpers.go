package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/alisaviation/pkg/logger"
)

//func (s *OrdersService) checkAccrualStatus(orderNumber string) {
//	for {
//		resp, err := s.AccrualClient.GetOrderInfo(orderNumber)
//		if err != nil {
//			logger.Log.Error("Failed to check accrual status",
//				zap.String("orderNumber", orderNumber),
//				zap.Error(err))
//			time.Sleep(5 * time.Second)
//			continue
//		}
//
//		if resp == nil {
//			time.Sleep(1 * time.Second)
//			continue
//		}
//
//		//// Обновляем статус в БД
//		//if err := s.OrderDB.UpdateOrderStatus(resp.Order, resp.Status, resp.Accrual); err != nil {
//		//	logger.Log.Error("Failed to update order status",
//		//		zap.String("orderNumber", resp.Order),
//		//		zap.Error(err))
//		//}
//		//
//		//// Если статус финальный - прекращаем проверку
//		//if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
//		//	break
//		//}
//
//		time.Sleep(1 * time.Second)
//	}
//}

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
	// Получаем заказы, которые ещё не в финальном статусе
	orders, err := s.OrderDB.GetOrdersByStatuses([]string{"NEW", "PROCESSING"})
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
			return
		default:
			accrualInfo, err := s.AccrualClient.GetOrderAccrual(ctx, order.Number)
			if err != nil {
				logger.Log.Warn("Failed to get accrual info for order",
					zap.String("order", order.Number),
					zap.Error(err))
				continue
			}

			if accrualInfo == nil {
				continue
			}

			if order.Status != accrualInfo.Status || order.Accrual != accrualInfo.Accrual {
				if err := s.OrderDB.UpdateOrderFromAccrual(
					order.Number,
					accrualInfo.Status,
					accrualInfo.Accrual,
				); err != nil {
					logger.Log.Error("Failed to update order status",
						zap.String("order", order.Number),
						zap.Error(err))
				} else {
					logger.Log.Info("Order status updated",
						zap.String("order", order.Number),
						zap.String("status", accrualInfo.Status),
						zap.Float64("accrual", accrualInfo.Accrual))
				}
			}
		}
	}
}
