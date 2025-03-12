package utils

import "github.com/rtmelsov/GopherMart/internal/models"

func Error(err error, code int) *models.Error {
	return &models.Error{Error: err.Error(), Code: code}

}
