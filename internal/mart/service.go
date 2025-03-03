package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	internalErrors "github.com/ruslantos/go-musthave-diploma-tpl/internal/errors"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/middlware/logger"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) error {
	err := s.repo.CreateUser(ctx, login, password)
	if err != nil {
		if err == internalErrors.ErrLoginAlreadyExists {
			return err
		}
		logger.Get().Error("Failed to register user", zap.Error(err))
		return errors.New("internal server error")
	}
	return nil
}

func (s *UserService) Authenticate(ctx context.Context, login, password string) bool {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil || user == nil {
		return false
	}
	return user.Password == password
}
