package routers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/auth"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/db"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRepo interface {
	CreateUser(login, passwordHash string) error
	IsLoginExist(login string) (bool, error)
	GetUserByLogin(login string) (*db.User, error)
}

func RegisterHandler(repo UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		req.Login = strings.TrimSpace(req.Login)
		if req.Login == "" || req.Password == "" {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		exists, err := repo.IsLoginExist(req.Login)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "логин уже занят", http.StatusConflict)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		if err := repo.CreateUser(req.Login, string(hash)); err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		token, err := auth.GenerateJWT(req.Login)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}

func LoginHandler(repo UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		req.Login = strings.TrimSpace(req.Login)
		if req.Login == "" || req.Password == "" {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		user, err := repo.GetUserByLogin(req.Login)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "неверная пара логин/пароль", http.StatusUnauthorized)
				return
			}
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
			http.Error(w, "неверная пара логин/пароль", http.StatusUnauthorized)
			return
		}
		token, err := auth.GenerateJWT(req.Login)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}
