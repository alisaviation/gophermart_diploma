package db

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
	"net/http"
)

func (db *DB) Register(value *models.DBUser) (*models.DBUser, *models.Error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	result := db.db.Create(&value)
	if result.Error != nil {
		return nil, db.ErrorHandler(result.Error.Error(), http.StatusConflict)
	}
	return value, nil
}

func (db *DB) Login(value *models.DBUser) (*models.DBUser, *models.Error) {
	var user models.DBUser
	db.mu.Lock()
	defer db.mu.Unlock()
	result := db.db.Where("login = ?", value.Login).First(&user)
	if result.Error != nil {
		return nil, db.ErrorHandler(result.Error.Error(), http.StatusUnauthorized)
	}

	if utils.CheckPassword(user.Password, value.Password) {
		return &user, nil
	}
	return nil, db.ErrorHandler("wrong password", http.StatusUnauthorized)
}

func (db *DB) GetUser(id uint) (*models.DBUser, *models.Error) {
	var user models.DBUser
	result := db.db.First(&user, id)
	if result.Error != nil {
		return nil, db.ErrorHandler(result.Error.Error(), http.StatusInternalServerError)
	}

	return &user, nil
}
