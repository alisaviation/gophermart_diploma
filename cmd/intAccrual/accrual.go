package intAccrual

import (
	"time"

	"go.uber.org/zap"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	storage "github.com/Tanya1515/gophermarket/cmd/storage"
)

type AccrualSystem struct {
	AccrualAddress string
	Storage        storage.StorageInterface
	Limit          int
	Logger         zap.SugaredLogger
}

func (ac *AccrualSystem) AccrualMain() {
	processOrderChan := make(chan add.OrderAcc, ac.Limit)
	orderIdChan := make(chan string, ac.Limit)
	resultOrderChan := make(chan add.OrderAcc, ac.Limit)

	go ac.Storage.StartProcessingUserOrder(ac.Logger, processOrderChan)

	go ac.SendOrder(processOrderChan, orderIdChan)

	go ac.GetOrderFromAccrual(orderIdChan, resultOrderChan)

	for order := range resultOrderChan {
		order := order

		go func() {
			for {
				var orderResult add.Order
				orderResult.Number = order.Order
				orderResult.Status = order.Status
				orderResult.Accrual = order.Accrual
				err := ac.Storage.ProcessAccOrder(orderResult)
				if err == nil {
					ac.Logger.Infof("Save recent information about order: %s", order.Order)
					break
				}
				time.Sleep(5 * time.Microsecond)
				ac.Logger.Errorf("Error while updating order %s to database: %s", order.Order, err)
			}
		}()
	}

}
