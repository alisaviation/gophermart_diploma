package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"

	handler "github.com/ruslantos/go-musthave-diploma-tpl/internal/handlers/register"
	service "github.com/ruslantos/go-musthave-diploma-tpl/internal/mart"
	authMiddlware "github.com/ruslantos/go-musthave-diploma-tpl/internal/middlware/auth"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/middlware/logger"
	"github.com/ruslantos/go-musthave-diploma-tpl/internal/repository"
)

func main() {
	// Инициализация логгера
	defer logger.Sync()

	var db *sqlx.DB
	db, err := sqlx.Open("pgx", "user=videos password=password dbname=shortenerdatabase sslmode=disable")
	if err != nil {
		logger.Get().Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Get().Fatal("Failed to ping database", zap.Error(err))
	}
	// Инициализация репозитория
	userRepo := repository.NewUserRepository(db)
	err = userRepo.InitStorage()

	// Инициализация сервиса
	userService := service.NewUserService(userRepo)

	// Инициализация хендлеров
	userHandler := handler.NewUserHandler(userService)

	r := chi.NewRouter()
	r.Use(authMiddlware.AuthMiddleware(userService))

	r.Post("/api/user/register", userHandler.Register)

	http.ListenAndServe(":8080", r)
}
