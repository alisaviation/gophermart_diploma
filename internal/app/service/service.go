package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/phedde/luhn-algorithm"
	"github.com/valinurovdenis/gomart/internal/app/accrualorder"
	"github.com/valinurovdenis/gomart/internal/app/currencybalance"
	"github.com/valinurovdenis/gomart/internal/app/orderstorage"
	"github.com/valinurovdenis/gomart/internal/app/userstorage"
	"github.com/valinurovdenis/gomart/internal/app/withdrawstorage"
)

type OrderService struct {
	UserStorage         userstorage.UserStorage
	UserBalanceStorage  userstorage.BalanceStorage
	WithdrawStorage     withdrawstorage.WithdrawStorage
	OrderStorage        orderstorage.OrderStorage
	AccrualOrderService accrualorder.AccrualOrderService
}

func (s *OrderService) orderIsValid(number string) bool {
	numberInt, err := strconv.ParseInt(string(number), 10, 64)
	if err != nil {
		return false
	}
	return luhn.IsValid(numberInt)
}

var ErrInvalidOrder = errors.New("invalid order")

func (s *OrderService) AddUserOrder(context context.Context, login string, number string) error {
	if !s.orderIsValid(number) {
		return ErrInvalidOrder
	}
	order, err := s.AccrualOrderService.GetOrder(context, login, number)
	if err != nil {
		return err
	}
	var balance currencybalance.CurrencyBalance
	balance.SetFloat(order.Accrual)
	userOrder := orderstorage.UserOrder{Login: login, Number: number, Status: order.Status, Balance: balance}
	err = s.OrderStorage.AddUserOrder(context, userOrder)
	if err == nil && !orderstorage.IsFinal(userOrder.Status) {
		err = s.AccrualOrderService.EnqueueOrderUpdate(context, login, number)
	}
	return err
}

func (s *OrderService) GetUserOrders(context context.Context, login string) ([]orderstorage.UserOrder, error) {
	return s.OrderStorage.GetUserOrders(context, login)
}

func (s *OrderService) GetUserBalance(context context.Context, login string) (userstorage.UserBalance, error) {
	return s.UserBalanceStorage.GetBalance(context, login)
}

var ErrNotEnoughBalance = errors.New("not enough balance for withdraw")

func (s *OrderService) AddUserWithdraw(context context.Context, withdraw withdrawstorage.UserWithdraw) error {
	if !s.orderIsValid(withdraw.Number) {
		return ErrInvalidOrder
	}

	userBalance, err := s.UserBalanceStorage.GetBalance(context, withdraw.Login)
	if err != nil {
		return err
	}
	if userBalance.Current.Less(withdraw.Withdraw) {
		return ErrNotEnoughBalance
	}
	err = s.WithdrawStorage.AddUserWithdraw(context, withdraw)
	return err
}

func (s *OrderService) GetUserWithdrawals(context context.Context, login string) ([]withdrawstorage.UserWithdraw, error) {
	return s.WithdrawStorage.GetUserWithdrawals(context, login)
}

func NewOrderService(userStorage userstorage.UserStorage,
	userBalanceStorage userstorage.BalanceStorage,
	withdrawStorage withdrawstorage.WithdrawStorage,
	orderStorage orderstorage.OrderStorage,
	accrualOrderService accrualorder.AccrualOrderService) *OrderService {

	ret := &OrderService{
		UserStorage:         userStorage,
		UserBalanceStorage:  userBalanceStorage,
		WithdrawStorage:     withdrawStorage,
		OrderStorage:        orderStorage,
		AccrualOrderService: accrualOrderService,
	}
	return ret
}
