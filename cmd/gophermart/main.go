package main

import (
	"log"
	"net/http"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/config"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/db"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/routers"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/service"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.New()

	dbConn, err := db.Init(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbConn.Close()

	if err := db.Migrate(dbConn); err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}

	userRepo := db.NewUserRepoPG(dbConn)
	userService := service.NewUserService(userRepo)
	orderRepo := db.NewOrderRepoPG(dbConn)
	orderService := service.NewOrderService(orderRepo, userRepo)
	h := routers.NewHandler(userService, orderService)
	r := routers.SetupRouters(h)

	log.Printf("Сервер запущен на %s", cfg.RunAddress)

	http.ListenAndServe(cfg.RunAddress, r)
}
