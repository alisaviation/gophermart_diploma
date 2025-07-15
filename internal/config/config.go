package config

import (
	"flag"
	"os"
)

type Server struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

func SetConfigServer() Server {
	var config Server

	setDefuaultConfig(config)
	setFlagsConfig(config)
	setEnvsConfig(config)

	return config
}

func setDefuaultConfig(config Server) {
	config.RunAddress = "localhost:8080"
	config.AccrualSystemAddress = "localhost:8080"
	config.DatabaseURI = "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"
	config.JWTSecret = "secret"
}

func setFlagsConfig(config Server) {
	address := flag.String("a", "localhost:8080", "HTTP server address")
	accrual := flag.String("r", "localhost:8080", "Accrual system address")
	database := flag.String("d", "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable", "Database URI")
	flag.Parse()
	config.RunAddress = *address
	config.AccrualSystemAddress = *accrual
	config.DatabaseURI = *database
}

func setEnvsConfig(config Server) {
	if envAddress := os.Getenv("RUN_ADDRESS"); envAddress != "" {
		config.RunAddress = envAddress
	}
	if envAccural := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccural != "" {
		config.AccrualSystemAddress = envAccural
	}
	if envDatabase := os.Getenv("DATABASE_URI"); envDatabase != "" {
		config.DatabaseURI = envDatabase
	}
}
