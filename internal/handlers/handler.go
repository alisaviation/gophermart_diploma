package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/config"
	"github.com/rtmelsov/GopherMart/internal/services"
)

type Handler struct {
	conf config.ConfigI
	serv services.ServiceI
}

type HandlerI interface {
	Login(c *gin.Context)
	Register(c *gin.Context)

	PostOrders(c *gin.Context)
	GetOrders(c *gin.Context)

	GetBalance(c *gin.Context)

	PostBalanceWithdraw(c *gin.Context)
	GetWithdrawals(c *gin.Context)
}

func NewHandler(conf config.ConfigI, service services.ServiceI) HandlerI {
	return &Handler{
		conf: conf,
		serv: service,
	}
}
