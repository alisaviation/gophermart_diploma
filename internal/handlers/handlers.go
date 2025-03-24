package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/middleware"
)

func SetupRouter(server *gin.Engine, cfg *config.Config) {
	server.Use(middleware.AuthMiddleware(cfg))
	server.POST("/api/user/register", func(c *gin.Context) { RegisterHandler(c, cfg) })
	server.POST("/api/user/login", func(c *gin.Context) { Login(c, cfg) })
	server.POST("/api/user/orders", func(c *gin.Context) { CreateOrder(c, cfg) })
	server.GET("/api/user/orders", func(c *gin.Context) { GetOrders(c, cfg) })
	server.GET("/api/user/balance")
	server.POST("/api/user/balance/withdraw")
	server.GET("/api/user/withdrawals")
}
