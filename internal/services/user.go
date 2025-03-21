package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/rtmelsov/GopherMart/internal/utils"
)

func (s *Service) Register(t *models.User) (*models.UserResponse, *models.Error) {
	password, err := utils.HashPassword(t.Password)
	if err != nil {
		return nil, err
	}
	t.Password = string(password)

	user, err := s.repo.Register(t)
	if err != nil {
		return nil, err
	}

	tokenString, err := utils.GetToken(user.ID, s.conf.GetEnvVariables().Secret)
	if err != nil {
		return nil, err
	}

	return &models.UserResponse{
		Message: "Ok",
		Token:   tokenString,
	}, nil
}

func (s *Service) Login(t *models.User) (*models.UserResponse, *models.Error) {
	user, err := s.repo.Login(t)
	if err != nil {
		return nil, err
	}

	tokenString, err := utils.GetToken(user.ID, s.conf.GetEnvVariables().Secret)
	if err != nil {
		return nil, err
	}

	return &models.UserResponse{
		Message: "Ok",
		Token:   tokenString,
	}, nil
}
