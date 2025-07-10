package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/middleware"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
)

// Router настраивает HTTP роутер
type Router struct {
	handlers *Handlers
	router   *chi.Mux
}

// NewRouter создает новый роутер
func NewRouter(storage storage.Storage, authService *services.AuthService, accrualService *services.AccrualService) *Router {
	handlers := NewHandlers(storage, authService, accrualService)
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.GzipMiddleware)

	// Публичные маршруты
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.RegisterHandler)
		r.Post("/login", handlers.LoginHandler)
	})

	// Защищенные маршруты
	router.Route("/api/user", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authService))

		r.Post("/orders", handlers.UploadOrderHandler)
		r.Get("/orders", handlers.GetOrdersHandler)
		r.Get("/balance", handlers.GetBalanceHandler)
		r.Post("/balance/withdraw", handlers.WithdrawHandler)
		r.Get("/withdrawals", handlers.GetWithdrawalsHandler)
	})

	return &Router{
		handlers: handlers,
		router:   router,
	}
}

// GetRouter возвращает настроенный роутер
func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
