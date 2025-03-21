package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handler) Register(c *gin.Context) {
	var req models.User
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp, localErr := h.serv.Register(&req)
	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", "Bearer "+resp.Token, 3600*24*30, "", "", false, true)
	c.Header("Content-Type", "text/plain; charset=utf-8")

	c.JSON(200, resp.Message)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.User
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}

	resp, localErr := h.serv.Login(&req)
	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", "Bearer "+resp.Token, 3600*24*30, "", "", false, true)
	c.Header("Content-Type", "text/plain; charset=utf-8")

	c.JSON(200, resp.Message)
}
