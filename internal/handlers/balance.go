package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) GetBalance(c *gin.Context) {
	id := c.GetUint("userId")
	balance, localErr := h.serv.GetBalance(&id)

	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(200, *balance)
}
