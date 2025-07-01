package services

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/alisaviation/internal/database"

	"github.com/alisaviation/internal/gophermart/models"
)

type AuthService interface {
	Register(login, password string) (string, error)
	//Login(login, password string) (string, error)
}

type authService struct {
	userRepo   database.UserRepository
	jwtService JWTServiceInterface
	//jwtService *JWTService
}

func NewAuthService(userRepo database.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: NewJWTService(jwtSecret, "gophermart"),
	}
}

func (s *authService) Register(login, password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if !utf8.ValidString(password) {
		return "", fmt.Errorf("password contains invalid UTF-8 sequences")
	}
	existingUser, err := s.userRepo.GetUserByLogin(login)
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
	if err := s.userRepo.CreateUser(user); err != nil {
		return "", fmt.Errorf("user creation failed: %w", err)
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

var ErrLoginTaken = errors.New("login already taken")
