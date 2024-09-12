package router

import (
	"context"
	"sync"
	"time"

	"GopherMart/internal/events"
)

func (s *serverMart) updateAccrual() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var wg, wgTimer sync.WaitGroup
	for {
		orders, err := s.DB.ReadAllOrderAccrualNoComplite(ctx)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if len(orders) != 0 {
			for _, order := range orders {
				wg.Add(1)
				go s.worker(order.Order, order.Login, &wg, &wgTimer)
			}
		}

		wg.Wait()
	}

}

func (s *serverMart) worker(order string, login string, wg, wgTimer *sync.WaitGroup) {
	wgTimer.Wait()
	accrual, sec, err := events.AccrualGet(s.Cfg.AccrualAddress, order)
	for sec != 0 {
		wgTimer.Add(1)

		wgTimer.Done()
		accrual, sec, err = events.AccrualGet(s.Cfg.AccrualAddress, order)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err != nil {
		return
	}
	_ = s.DB.UpdateOrderAccrual(ctx, login, accrual)
	wg.Done()
}
