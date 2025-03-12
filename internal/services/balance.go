package services

import "github.com/rtmelsov/GopherMart/internal/models"

func (s *Service) GetBalance(id *uint) (*models.DBBalance, *models.Error) {
	return s.repo.GetBalance(id)
}
