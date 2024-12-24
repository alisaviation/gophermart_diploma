package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/valinurovdenis/gomart/internal/app/userstorage"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

const tokenExpiration = time.Hour * 3

type JwtAuthenticator struct {
	SecretKey   string
	UserStorage userstorage.UserStorage
}

func NewAuthenticator(secretKey string, userStorage userstorage.UserStorage) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey:   secretKey,
		UserStorage: userStorage,
	}
}

func (a *JwtAuthenticator) buildJWTString(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *JwtAuthenticator) getLogin(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.SecretKey), nil
		})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return claims.Login, nil
}

func (a *JwtAuthenticator) setCookie(w http.ResponseWriter, login string) {
	token, _ := a.buildJWTString(login)
	newCookie := http.Cookie{Name: "Authorization", Value: token}
	http.SetCookie(w, &newCookie)
	w.WriteHeader(http.StatusOK)
}

func (a *JwtAuthenticator) Register(w http.ResponseWriter, r *http.Request) {
	var loginPassword userstorage.LoginPassword
	if err := json.NewDecoder(r.Body).Decode(&loginPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.UserStorage.AddUser(r.Context(), loginPassword); err != nil {
		if errors.Is(err, userstorage.ErrLoginExists) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	a.setCookie(w, loginPassword.Login)
}

func (a *JwtAuthenticator) Login(w http.ResponseWriter, r *http.Request) {
	var loginPassword userstorage.LoginPassword
	if err := json.NewDecoder(r.Body).Decode(&loginPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password, err := a.UserStorage.GetUserPassword(r.Context(), loginPassword.Login)
	if err != nil || password != loginPassword.Password {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	a.setCookie(w, loginPassword.Login)
}

func (a *JwtAuthenticator) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		var login string
		if err == nil {
			login, err = a.getLogin(cookie.Value)
		}

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Set("Login", login)

		h.ServeHTTP(w, r)
	})
}
