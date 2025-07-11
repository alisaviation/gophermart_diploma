package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/config"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/server"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	dbStorage, err := storage.NewDatabaseStorage(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStorage.Close()

	// Генерируем секретный ключ для JWT
	jwtSecret, err := services.GenerateSecret()
	if err != nil {
		log.Fatalf("Failed to generate JWT secret: %v", err)
	}

	// Создаем сервисы
	authService := services.NewAuthService(jwtSecret)
	accrualService := services.NewAccrualService(cfg.AccrualSystemAddress)

	// Создаем роутер
	router := server.NewRouter(dbStorage, authService, accrualService)

	orderProcessInterval, err := cfg.GetOrderProcessInterval()
	if err != nil {
		log.Fatalf("Failed to parse order process interval: %v", err)
	}

	// Создаем процессор заказов
	orderProcessor := server.NewOrderProcessor(dbStorage, accrualService, orderProcessInterval)
	orderProcessor.Start()
	defer orderProcessor.Stop()

	// HTTP сервер
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router.GetRouter(),
	}

	// Обработкой сигналов для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Starting server on %s", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
