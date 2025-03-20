package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (h *Handler) PostOrders(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.conf.GetLogger().Error("try to get body from post orders", zap.Error(err))
		c.String(http.StatusUnprocessableEntity, "Failed to read body")
		return
	}

	id := c.GetUint("userId")

	dbRequest := &models.DBOrder{
		UserID: id,
		Number: string(body),
	}

	h.conf.GetLogger().Info("POST REQUEST --START", zap.String("ORDER NUMBER", dbRequest.Number))

	localErr := h.serv.PostOrders(dbRequest)

	if localErr != nil {
		h.conf.GetLogger().Info("POST REQUEST END WITH ERROR", zap.String("ERROR TEXT", localErr.Error))
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	h.conf.GetLogger().Info("POST REQUEST END WITH SUCCESS")
	c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) GetOrders(c *gin.Context) {
	id := c.GetUint("userId")
	list, localErr := h.serv.GetOrders(&id)
	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, *list)
}
