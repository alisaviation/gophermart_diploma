package intAccrual

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

// ограниение по количеству запросов? (либо через цикл, либо через семафоры)
func (ac *AccrualSystem) SendOrder(inputChan chan add.OrderAcc, resultChan chan string) {

	for order := range inputChan {
		order := order

		go func() {
			client := resty.New()

			ordersByte, err := json.Marshal(order)
			if err != nil {
				ac.Logger.Errorf("Error while marshalling order %s: %s", order.Order, err)
			}
			for {
				_, err = client.R().SetHeader("Content-Type", "application/json").
					SetBody(ordersByte).
					Post("http://" + ac.AccrualAddress + "/api/orders")

				if err == nil {
					ac.Logger.Infof("Send order: %s", order.Order)
					break
				}

				time.Sleep(5 * time.Microsecond)
				ac.Logger.Errorf("Error while sending order %s to accrual system: %s", order.Order, err)

			}

			resultChan <- order.Order
		}()

	}
}

// ограничение запросов (семафоры)
func (ac *AccrualSystem) GetOrderFromAccrual(inputChan chan string, resultChan chan add.OrderAcc) {

	var order add.OrderAcc

	for orderId := range inputChan {
		orderId := orderId

		go func() {
			client := resty.New()

			for {
				resp, err := client.R().Get("http://" + ac.AccrualAddress + "/api/orders/" + orderId)

				if err != nil {
					time.Sleep(5 * time.Microsecond)
					ac.Logger.Errorf("Error while getting order %s from accrual system: %s", orderId, err)
					continue
				}

				err = json.Unmarshal(resp.Body(), &order)
				if err != nil {
					// ac.Logger.Errorf("Error while unmarshalling order %s: %s", order.Order, err)
				}

				if (order.Status == "PROCESSED") || (order.Status == "INVALID") {
					resultChan <- order
					break
				}
			}

		}()

	}
}
