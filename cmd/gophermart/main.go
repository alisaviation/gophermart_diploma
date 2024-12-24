package main

import (
	"database/sql"
	"errors"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/valinurovdenis/gomart/internal/app/accrualorder"
	"github.com/valinurovdenis/gomart/internal/app/auth"
	"github.com/valinurovdenis/gomart/internal/app/handlers"
	"github.com/valinurovdenis/gomart/internal/app/logger"
	"github.com/valinurovdenis/gomart/internal/app/orderstorage"
	"github.com/valinurovdenis/gomart/internal/app/service"
	"github.com/valinurovdenis/gomart/internal/app/userstorage"
	"github.com/valinurovdenis/gomart/internal/app/withdrawstorage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	config := new(Config)
	parseFlags(config)
	config.updateFromEnv()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	if config.DatabaseURI == "" {
		return errors.New("empty database config")
	}
	db, err := sql.Open("pgx", config.DatabaseURI)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userStorage := userstorage.NewDatabaseUserStorage(db)
	withdrawStorage := withdrawstorage.NewDatabaseWithdrawStorage(db)
	orderStorage := orderstorage.NewDatabaseOrderStorage(db)
	accrualOrderService := accrualorder.NewAccrualOrderQueue(db, 10, config.AccrualSystemAddress, orderStorage)
	auth := auth.NewAuthenticator(config.SecretKey, userStorage)
	service := service.NewOrderService(userStorage, userStorage, withdrawStorage, orderStorage, accrualOrderService)
	handler := handlers.NewMartHandler(*service, *auth)

	return http.ListenAndServe(config.RunAddress, handlers.MartRouter(*handler))
}
