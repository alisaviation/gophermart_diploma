package utils

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HashPassword(password string) ([]byte, *models.Error) {
	res, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, Error(err, http.StatusInternalServerError)
	}
	return res, nil
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
