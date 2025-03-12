package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (h *Handler) PostOrders(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read body")
		return
	}

	// Convert string to int64
	num, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(500, err)
		return
	}

	id := c.GetUint("userId")

	dbRequest := &models.DBOrder{
		UserID: id,
		Number: num,
	}

	localErr := h.serv.PostOrders(dbRequest)
	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}
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
	c.JSON(200, *list)
}
