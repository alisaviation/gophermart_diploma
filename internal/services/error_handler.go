package services

import (
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
)

func (s *Service) ErrorHandler(err string, code int) *models.Error {
	s.conf.GetLogger().Error("DB ERROR HANDLER", zap.String("ERROR TEXT", err), zap.Int("ERROR CODE", code))
	return &models.Error{
		Error: err,
		Code:  code,
	}
}
