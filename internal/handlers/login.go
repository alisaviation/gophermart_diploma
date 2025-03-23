package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/storage"
)

func Login(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var infoUser AuthForm

	if err := c.ShouldBindJSON(&infoUser); err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, "Пустой body")
		return
	}

	if infoUser.Login == "" || infoUser.Password == "" {
		c.JSON(http.StatusBadRequest, "PRoblem with JSON")
		return
	}

	infoUser.Password = HashPassword(cfg, infoUser.Password)

	ctx := c.Request.Context()
	passwordUser, err := cfg.Store.GetPasswordFromLogin(ctx, infoUser.Login)

	if err != nil {
		cfg.Sugar.Error(err)
		if errors.Is(err, storage.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, "user is not found")
			return
		}
		c.JSON(http.StatusBadGateway, "Problem with service")
		return
	}

	if passwordUser != infoUser.Password {
		c.JSON(http.StatusUnauthorized, "Password is not correct")
		return
	}

	id, err := cfg.Store.GetID(ctx, infoUser.Login)

	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Problem with service")
		return
	}

	token, err := jwtauth.BuildJWTString(id)

	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Problem With Service")
		return
	}

	c.SetCookie("jwt", token, 3600, "/", "", false, false)
	jwtHeader := fmt.Sprintf("Bearer %v", token)

	c.Writer.Header().Set("Authorization", jwtHeader)
	c.JSON(http.StatusOK, "User auth successfully")
}
