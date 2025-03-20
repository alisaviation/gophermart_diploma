package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/config"
	"github.com/rtmelsov/GopherMart/internal/db"
	"github.com/rtmelsov/GopherMart/internal/handlers"
	"github.com/rtmelsov/GopherMart/internal/middleware"
	"github.com/rtmelsov/GopherMart/internal/repository"
	"github.com/rtmelsov/GopherMart/internal/services"
)

func main() {
	// получили все перемененные окружения
	c, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	g := gin.Default()

	// подключаемся к базе данных
	d, err := db.GetDb(c)
	if err != nil {
		panic(err)
	}

	// передаем методы базы данных в репозитории чтобы мы могли работать с ним через него (изолируем данные базы данных кроме методов)
	r := repository.GetRepository(c, d)

	// передаем методы репозитории в сервисы чтобы к ним обращаться через них, а не напрямую (изолируем данные репозитории кроме методов)
	s := services.NewService(c, r)

	// передаем в ручки все методы сервиса (изолируем данные сервиса кроме методов)
	h := handlers.NewHandler(c, s)

	// middleware
	g.Use(gin.Logger())

	// добавили в начале /api/user
	clientApi := g.Group(c.GetEnvVariables().RootUrl)

	clientApi.POST("/register", h.Register)
	clientApi.POST("/login", h.Login)

	protected := g.Group(c.GetEnvVariables().RootUrl)
	protected.Use(middleware.Auth(c, d))
	{
		protected.POST("/orders", h.PostOrders)
		protected.GET("/orders", h.GetOrders)

		protected.GET("/balance", h.GetBalance)

		protected.POST("/balance/withdraw", h.PostBalanceWithdraw)
		protected.GET("/withdrawals", h.GetWithdrawals)
	}

	if err := g.Run(c.GetEnvVariables().RunAddress); err != nil {
		return
	}
}
