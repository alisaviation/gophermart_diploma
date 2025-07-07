package services

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/alisaviation/internal/database"

	"github.com/alisaviation/internal/gophermart/models"
)

var (
	ErrLoginTaken         = errors.New("login already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

type AuthStructService struct {
	UserRepo   database.User
	JwtService JWTServiceInterface
}

func NewAuthService(userRepo database.User, jwtSecret string) AuthService {
	return &AuthStructService{
		UserRepo:   userRepo,
		JwtService: NewJWTService(jwtSecret, "gophermart"),
	}
}

func (s *AuthStructService) Register(login, password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if !utf8.ValidString(password) {
		return "", fmt.Errorf("password contains invalid UTF-8 sequences")
	}
	existingUser, err := s.UserRepo.GetUserByLogin(login)
	if err != nil {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return "", ErrLoginTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}

	user := models.User{
		Login:        login,
		PasswordHash: string(hashedPassword),
	}
	if err := s.UserRepo.CreateUser(user); err != nil {
		return "", fmt.Errorf("user creation failed: %w", err)
	}

	token, err := s.JwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *AuthStructService) Login(login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.UserRepo.GetUserByLogin(login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	token, err := s.JwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
