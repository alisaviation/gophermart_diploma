package config

import (
	"fmt"

	"github.com/skakunma/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
)

type Config struct {
	Store         storage.PostgresStorage
	Salt          string
	Sugar         *zap.SugaredLogger
	FlagForDB     string
	FlagAddress   string
	FlagAddressAS string
}

func NewConfig() (*Config, error) {
	storage, err := storage.CreatePostgreStorage("host=localhost user=postgres password=example dbname=diplomka sslmode=disable")
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации логгера: %w", err)
	}

	salt := "random_salt_123"

	cfg := &Config{Store: *storage, Salt: salt, Sugar: logger.Sugar()}

	ParsePlags(cfg)

	return cfg, nil
}
