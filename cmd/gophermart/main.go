package main

import (
	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/handlers"
)

func main() {
	cfg, err := config.NewConfig()

	if err != nil {
		panic(err)
	}

	server := gin.Default()

	handlers.SetupRouter(server, cfg)

	server.Run(cfg.FlagAddress)
}
