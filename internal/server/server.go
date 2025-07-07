package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alisaviation/internal/database"
	"github.com/alisaviation/internal/database/postgres"
	"github.com/alisaviation/internal/gophermart/services"
	"github.com/alisaviation/internal/handlers"
	"github.com/alisaviation/internal/middleware"

	"go.uber.org/zap"

	"github.com/alisaviation/internal/config"
	"github.com/alisaviation/pkg/logger"
)

type ServerApp struct {
	config         config.Server
	httpServer     *http.Server
	shutdownSignal chan struct{}
	db             *sql.DB
	storage        database.Storage
	wg             sync.WaitGroup
	mu             sync.RWMutex
	jwtSecret      string
}

func NewServerApp(conf config.Server) *ServerApp {
	return &ServerApp{
		config:         conf,
		shutdownSignal: make(chan struct{}),
	}
}

func (s *ServerApp) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := s.initDB(ctx)
	if err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	s.storage = storage

	go s.handleSignals(cancel)

	if err := s.startHTTPServer(); err != nil {
		return err
	}

	select {
	case <-s.shutdownSignal:
		logger.Log.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Log.Info("Context cancelled")
	}
	s.shutdown(ctx)
	logger.Log.Info("Server shutdown complete")
	return nil
}

func (s *ServerApp) handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Log.Info("Received signal", zap.String("signal", sig.String()))
	cancel()
	close(s.shutdownSignal)
}

func (s *ServerApp) startHTTPServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r := chi.NewRouter()

	r.Use(
		logger.RequestResponseLogger,
		middleware.GzipMiddleware,
	)

	s.registerRoutes(r)

	s.httpServer = &http.Server{
		Addr:    s.config.RunAddress,
		Handler: r,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		logger.Log.Info("Starting HTTP server", zap.String("address", s.config.RunAddress))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("HTTP server failed", zap.Error(err))
		}
	}()

	return nil
}
func (s *ServerApp) registerRoutes(r *chi.Mux) {
	jwtService := services.NewJWTService(s.jwtSecret, "gophermart")
	authService := services.NewAuthService(s.storage, s.jwtSecret)
	authHandler := handlers.NewAuthHandler(authService)
	accrualClient := services.NewAccrualClient(s.config.AccrualSystemAddress)
	orderService := services.NewOrderService(s.storage, accrualClient)
	orderHandler := handlers.NewOrderHandler(orderService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go orderService.StartStatusUpdater(ctx, 1*time.Minute)
	//balance := handlers.NewBalanceController(
	//	handlers.NewBalanceService(s.balanceRepo))

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))

		r.Post("/api/user/orders", orderHandler.UploadOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)
		//	r.Get("/api/user/balance", s.balanceController.GetBalance)
		//	r.Post("/api/user/balance/withdraw", s.balanceController.Withdraw)
		//	r.Get("/api/user/withdrawals", s.balanceController.GetWithdrawals)
	})
}

func (s *ServerApp) shutdown(ctx context.Context) {
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("HTTP server shutdown failed", zap.Error(err))
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Log.Error("Failed to close database connection", zap.Error(err))
		}
	}

	s.wg.Wait()
}

func (s *ServerApp) initDB(ctx context.Context) (*postgres.PostgresStorage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.DatabaseURI == "" {
		return nil, errors.New("database URI not configured")
	}

	logger.Log.Info("Connecting to db", zap.String("DB URI", s.config.DatabaseURI))

	db, err := sql.Open("postgres", s.config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		db.Close()
		return nil, err
	}

	storage, err := postgres.NewPostgresDatabase(db)
	if err != nil {
		logger.Log.Fatal("Failed to create Postgres storage", zap.Error(err))
		db.Close()
		return nil, err
	}

	s.db = db
	logger.Log.Info("Successfully connected to database")

	return storage, nil
}
