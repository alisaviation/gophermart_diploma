package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
)

type Order struct {
	Number   int       `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual,omitempty"`
	Uploaded time.Time `json:"uploaded_at"`
}

type OrderInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"` // omitempty, чтобы не включать, если нет начислений
}

func isValidLuhn(number int) bool {
	str := strconv.Itoa(number)
	n := len(str)
	sum := 0
	isSecond := false

	for i := n - 1; i >= 0; i-- {
		digit := int(str[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

func CreateOrder(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") {
		c.JSON(http.StatusBadRequest, "Content-type must be text/plain")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, "Problem with parsing body")
		return
	}
	idOrder, err := strconv.Atoi(string(body))
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusUnprocessableEntity, "Order is cat'n be int")
	}
	if !isValidLuhn(idOrder) {
		c.JSON(http.StatusUnprocessableEntity, "Order is not Lun")
		return
	}

	user, _ := c.Get("user")
	claims, ok := user.(*jwtauth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	ctx := c.Request.Context()

	exist, err := cfg.Store.IsOrderExists(ctx, idOrder)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Error")
	}
	if exist {
		autgorID, err := cfg.Store.GetAuthorOrder(ctx, idOrder)
		if err != nil {
			cfg.Sugar.Error(err)
			c.JSON(http.StatusBadGateway, "Problem")
			return
		}

		if claims.UserID != autgorID {
			c.JSON(http.StatusConflict, "Order is was loaded")
			return
		}
		c.JSON(http.StatusOK, "Order is was loaded")
	}
	err = cfg.Store.CreateOrder(ctx, claims.UserID, idOrder)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Problem with storage")
		return
	}

	c.JSON(http.StatusAccepted, "was accepted")
}

func GetOrders(c *gin.Context, cfg *config.Config) {
	user, _ := c.Get("user")
	claims := user.(*jwtauth.Claims)
	userID := claims.UserID
	ctx := c.Request.Context()

	// Получаем список заказов пользователя
	orders, err := cfg.Store.GetOrdersFromUser(ctx, userID)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, "Id is a bad")
		return
	}

	// Расширяем заказы данными о начислениях
	var responseOrders []map[string]interface{}
	for _, order := range orders {
		orderInfo, err := GetInfoAboutOrder(ctx, cfg, order)
		if err != nil {
			cfg.Sugar.Warnf("Не удалось получить информацию о заказе %s: %v", order, err)
			continue
		}

		// Формируем ответ
		orderData := map[string]interface{}{
			"number":      order,
			"uploaded_at": time.Now(),
			"status":      "NEW", // По умолчанию, если нет данных от сервиса
		}
		if orderInfo != nil {
			orderData["status"] = orderInfo.Status
			if orderInfo.Accrual > 0 {
				orderData["accrual"] = orderInfo.Accrual
			}
		}

		responseOrders = append(responseOrders, orderData)
	}

	if len(responseOrders) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, responseOrders)
}

func GetInfoAboutOrder(ctx context.Context, cfg *config.Config, orderID int) (*OrderInfo, error) {
	address := fmt.Sprintf("%v/api/orders/%v", cfg.FlagAddressAS, orderID)
	fmt.Println(address)
	// HTTP-клиент с таймаутом
	client := &http.Client{Timeout: 5 * time.Second}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// Обрабатываем коды ответа
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Декодируем JSON-ответ
	var orderInfo OrderInfo
	if err := json.NewDecoder(resp.Body).Decode(&orderInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &orderInfo, nil
}
