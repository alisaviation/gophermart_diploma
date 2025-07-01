package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTServiceInterface interface {
	GenerateToken(userID int, login string) (string, error)
}

type JWTService struct {
	secretKey []byte
	issuer    string
}

func NewJWTService(secret string, issuer string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret),
		issuer:    issuer,
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
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
