package accrualorder

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/valinurovdenis/gomart/internal/app/currencybalance"
	"github.com/valinurovdenis/gomart/internal/app/orderstorage"
	"go.dataddo.com/pgq"
	"go.dataddo.com/pgq/x/schema"
)

const queueName = "orders_updater"

type AccrualOrder struct {
	Order   string                   `json:"order"`
	Status  orderstorage.OrderStatus `json:"status"`
	Accrual float64                  `json:"accrual"`
}

//go:generate mockery --name AccrualOrderService
type AccrualOrderService interface {
	GetOrder(context context.Context, login string, number string) (AccrualOrder, error)

	EnqueueOrderUpdate(context context.Context, login string, number string) error
}

type AccrualOrderQueue struct {
	DB                *sql.DB
	UpdateThreads     int
	AccrualServiceUrl string
	OrderStorage      orderstorage.OrderStorage
	Stop              func()
}

func (s *AccrualOrderQueue) Init() {
	tx, _ := s.DB.BeginTx(context.Background(), nil)
	create := schema.GenerateCreateTableQuery(queueName)
	tx.Exec(create)
	tx.Commit()
}

var ErrNoSuchOrder = errors.New("no such order in accrual service")
var ErrNoAnswer = errors.New("no answer from accrual service")

type QueueOrder struct {
	Login  string `json:"login"`
	Number string `json:"number"`
}

func (s *AccrualOrderQueue) getAccrualOrder(ctx context.Context, number string) (AccrualOrder, error) {
	var (
		err        error
		response   *http.Response
		retries    int           = 3
		retryDelay time.Duration = 500 * time.Millisecond
		backoff    int           = 2
	)
	for retries > 0 {
		response, err = http.Get(s.AccrualServiceUrl + fmt.Sprintf("/api/orders/%s", number))
		if err == nil && response.StatusCode != http.StatusTooManyRequests {
			if response.StatusCode == http.StatusNoContent {
				return AccrualOrder{}, ErrNoSuchOrder
			}

			var order AccrualOrder
			defer response.Body.Close()
			if err = json.NewDecoder(response.Body).Decode(&order); err != nil {
				return AccrualOrder{}, err
			}
			return order, nil
		} else {
			retries--
		}
		time.Sleep(retryDelay)
		retryDelay *= time.Duration(backoff)
	}
	return AccrualOrder{}, ErrNoAnswer
}

func (s *AccrualOrderQueue) EnqueueOrderUpdate(ctx context.Context, login string, number string) error {
	publisher := pgq.NewPublisher(s.DB)
	payload, _ := json.Marshal(QueueOrder{Login: login, Number: number})
	timeInFiveMinute := time.Now().Add(5 * time.Minute)
	msg := &pgq.MessageOutgoing{Payload: payload, ScheduledFor: &timeInFiveMinute, Metadata: map[string]string{"login": login, "number": number}}
	_, err := publisher.Publish(ctx, queueName, msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccrualOrderQueue) GetOrder(ctx context.Context, login string, number string) (AccrualOrder, error) {
	order, err := s.getAccrualOrder(ctx, number)
	if err != nil {
		return AccrualOrder{}, err
	}
	if order.Status == orderstorage.Registered {
		order.Status = orderstorage.New
	}
	return order, nil
}

func (s AccrualOrderQueue) HandleMessage(ctx context.Context, msg *pgq.MessageIncoming) (bool, error) {
	var queueOrder QueueOrder
	json.Unmarshal(msg.Payload, &queueOrder)
	order, err := s.getAccrualOrder(ctx, queueOrder.Number)
	if err != nil {
		return false, err
	}
	if !orderstorage.IsFinal(order.Status) {
		s.EnqueueOrderUpdate(ctx, queueOrder.Login, order.Order)
		return true, nil
	}
	var balance currencybalance.CurrencyBalance
	balance.SetFloat(order.Accrual)
	s.OrderStorage.UpdateProcessedOrder(ctx,
		orderstorage.UserOrder{
			Login:   queueOrder.Login,
			Balance: balance,
			Number:  order.Order,
			Status:  order.Status,
		})
	return true, nil
}

func (s AccrualOrderQueue) runUpdateThread(ctx context.Context) error {
	consumer, err := pgq.NewConsumer(s.DB, queueName, s)
	if err != nil {
		return err
	}

	err = consumer.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccrualOrderQueue) runBackgroundUpdate(ctx context.Context) {
	for range s.UpdateThreads {
		go s.runUpdateThread(ctx)
	}
}

func NewAccrualOrderQueue(db *sql.DB, updateThreads int, accrualServiceUrl string, orderStorage orderstorage.OrderStorage) *AccrualOrderQueue {
	ctx, stop := context.WithCancel(context.Background())
	ret := &AccrualOrderQueue{DB: db, UpdateThreads: updateThreads, AccrualServiceUrl: accrualServiceUrl, OrderStorage: orderStorage, Stop: stop}
	ret.Init()
	ret.runBackgroundUpdate(ctx)
	return ret
}
