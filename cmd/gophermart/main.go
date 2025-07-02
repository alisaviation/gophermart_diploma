package main

import (
	"log"
	"net/http"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/config"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/db"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/routers"
	"github.com/go-chi/chi/v5"
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

	r := chi.NewRouter()

	repo := db.NewUserRepoPG(dbConn)
	r.Post("/api/user/register", routers.RegisterHandler(repo))
	r.Post("/api/user/login", routers.LoginHandler(repo))

	log.Printf("Сервер запущен на %s", cfg.RunAddress)
	http.ListenAndServe(cfg.RunAddress, r)
}
