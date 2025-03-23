package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
)

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
