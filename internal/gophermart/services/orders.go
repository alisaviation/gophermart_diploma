package services

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/database/postgres"
	"github.com/alisaviation/internal/gophermart/models"
)

type OrderService interface {
	UploadOrder(userID int, orderNumber string) (int, error)
}

type orderService struct {
	orderDB database.Order
}

func NewOrderService(orderDB database.Order) OrderService {
	return &orderService{orderDB: orderDB}
}

func (s *orderService) ValidateOrderNumber(number string) bool {
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

func (s *orderService) UploadOrder(userID int, orderNumber string) (int, error) {
	if _, err := strconv.Atoi(orderNumber); err != nil {
		return http.StatusBadRequest, errors.New("order number must contain only digits")
	}

	if !s.ValidateOrderNumber(orderNumber) {
		return http.StatusUnprocessableEntity, errors.New("invalid order number by Luhn algorithm")
	}

	existingOrder, err := s.orderDB.GetOrderByNumber(orderNumber)
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

	if err := s.orderDB.CreateOrder(order); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
}
