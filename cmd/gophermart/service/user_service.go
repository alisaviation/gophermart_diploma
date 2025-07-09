package service

import (
	"errors"
	"strings"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/auth"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(login, passwordHash string) error
	IsLoginExist(login string) (bool, error)
	GetUserByLogin(login string) (*models.User, error)
}

type UserService struct {
	UserRepo UserRepo
}

var (
	ErrUserExists      = errors.New("логин уже занят")
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrInvalidPassword = errors.New("неверная пара логин/пароль")
)

func NewUserService(repo UserRepo) *UserService {
	return &UserService{UserRepo: repo}
}

func (s *UserService) Register(req models.RegisterRequest) (string, error) {
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		return "", errors.New("неверный формат запроса")
	}
	exists, err := s.UserRepo.IsLoginExist(req.Login)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := s.UserRepo.CreateUser(req.Login, string(hash)); err != nil {
		return "", err
	}
	token, err := auth.GenerateJWT(req.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) Login(req models.RegisterRequest) (string, error) {
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		return "", errors.New("неверный формат запроса")
	}
	user, err := s.UserRepo.GetUserByLogin(req.Login)
	if err != nil {
		return "", ErrUserNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return "", ErrInvalidPassword
	}
	token, err := auth.GenerateJWT(req.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) GetUserByLogin(login string) (*models.User, error) {
	return s.UserRepo.GetUserByLogin(login)
}
