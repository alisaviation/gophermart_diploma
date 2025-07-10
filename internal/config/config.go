package config

import (
	"flag"
	"fmt"
	"os"
)

// Config содержит конфигурацию сервера
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

// Load загружает конфигурацию из флагов и переменных окружения
func Load() (*Config, error) {
	var (
		flagRunAddress           string
		flagDatabaseURI          string
		flagAccrualSystemAddress string
	)

	flag.StringVar(&flagRunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagDatabaseURI, "d", "", "database URI")
	flag.StringVar(&flagAccrualSystemAddress, "r", "", "accrual system address")
	flag.Parse()

	// Приоритет: env > flag > default
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		flagRunAddress = envRunAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		flagDatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		flagAccrualSystemAddress = envAccrualSystemAddress
	}

	if flagDatabaseURI == "" {
		return nil, fmt.Errorf("DATABASE_URI is required")
	}

	return &Config{
		RunAddress:           flagRunAddress,
		DatabaseURI:          flagDatabaseURI,
		AccrualSystemAddress: flagAccrualSystemAddress,
	}, nil
}
