package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"net/http"
)

type Config struct {
	Logger       *zap.Logger
	EnvVariables *models.EnvVariables
}

type ConfigI interface {
	GetLogger() *zap.Logger
	GetEnvVariables() *models.EnvVariables
}

func (c Config) GetLogger() *zap.Logger {
	return c.Logger
}

func (c Config) GetEnvVariables() *models.EnvVariables {
	return c.EnvVariables
}

func InitConfig() (ConfigI, *models.Error) {
	var envVar models.EnvVariables
	flag.StringVar(&envVar.RunAddress, "a", "", "host and port to run services")
	flag.StringVar(&envVar.DataBaseURL, "d", "", "data base url to rub db")
	flag.StringVar(&envVar.AccuralSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&envVar.RootUrl, "u", "/api/user", "root url")
	flag.StringVar(&envVar.Secret, "s", "secret12345", "secret key")

	err := env.Parse(&envVar)
	if err != nil {
		return nil, &models.Error{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	flag.Parse()

	Log, _ := zap.NewProduction()
	return Config{
		Logger:       Log,
		EnvVariables: &envVar,
	}, nil
}
