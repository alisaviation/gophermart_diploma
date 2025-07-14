package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	SecretKey []byte
	Issuer    string
}

func NewJWTService(secret string, issuer string) *JWTService {
	return &JWTService{
		SecretKey: []byte(secret),
		Issuer:    issuer,
	}
}

type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateToken(userID int, login string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.SecretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.SecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	if claims.Issuer != s.Issuer {
		return nil, errors.New("invalid token Issuer")
	}

	return claims, nil
}
