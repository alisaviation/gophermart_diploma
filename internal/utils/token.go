package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
	"time"
)

func GetToken(id uint, secret string) (string, *models.Error) {
	var secretKey = []byte(secret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": id,
			"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", &models.Error{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	return tokenString, nil
}
