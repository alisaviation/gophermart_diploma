package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) GetBalance(c *gin.Context) {
	id := c.GetUint("userId")
	balance, localErr := h.serv.GetBalance(&id)

	h.conf.GetLogger().Info("balance info", zap.Float64("balance.Withdraw", balance.Withdraw), zap.Float64("balance.Current", balance.Current))

	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(200, *balance)
}
