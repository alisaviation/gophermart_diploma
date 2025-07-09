package service

import (
	"errors"
	"strings"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepo interface {
	CreateOrder(orderNumber string, userID int64) error
	GetOrderByNumber(orderNumber string) (*models.Order, error)
	GetOrderByNumberAndUserID(orderNumber string, userID int64) (*models.Order, error)
}

type OrderService struct {
	OrderRepo OrderRepo
	UserRepo  UserRepo
}

func NewOrderService(orderRepo OrderRepo, userRepo UserRepo) *OrderService {
	return &OrderService{OrderRepo: orderRepo, UserRepo: userRepo}
}

var (
	ErrOrderAlreadyUploadedByUser    = errors.New("order already uploaded by this user")
	ErrOrderAlreadyUploadedByAnother = errors.New("order already uploaded by another user")
	ErrInvalidOrderFormat            = errors.New("invalid order format")
	ErrInvalidOrderNumber            = errors.New("invalid order number")
)

func isValidLuhn(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if double {
			digit = digit * 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func (s *OrderService) UploadOrder(orderNumber string, userID int64) (int, error) {
	orderNumber = strings.TrimSpace(orderNumber)
	if orderNumber == "" {
		return 400, ErrInvalidOrderFormat
	}
	for _, c := range orderNumber {
		if c < '0' || c > '9' {
			return 422, ErrInvalidOrderNumber
		}
	}
	if !isValidLuhn(orderNumber) {
		return 422, ErrInvalidOrderNumber
	}
	order, err := s.OrderRepo.GetOrderByNumber(orderNumber)
	if err == nil && order != nil {
		if order.UserID == userID {
			return 200, ErrOrderAlreadyUploadedByUser
		} else {
			return 409, ErrOrderAlreadyUploadedByAnother
		}
	}
	err = s.OrderRepo.CreateOrder(orderNumber, userID)
	if err != nil {
		return 500, err
	}
	return 202, nil
}
