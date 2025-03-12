package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/models"
	"go.uber.org/zap"
)

func (h *Handler) PostBalanceWithdraw(c *gin.Context) {
	var req models.Withdrawal
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(500, err)
		return
	}

	id := c.GetUint("userId")
	DBReq := models.DBWithdrawal{
		Order:  req.Order,
		Sum:    req.Sum,
		UserID: id,
	}

	localErr := h.serv.PostBalanceWithdraw(&DBReq)
	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}
}

func (h *Handler) GetWithdrawals(c *gin.Context) {
	id := c.GetUint("userId")
	list, localErr := h.serv.GetWithdrawals(&id)

	if localErr != nil {
		h.conf.GetLogger().Error("error", zap.String("error", localErr.Error))
		c.JSON(localErr.Code, localErr.Error)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(200, *list)
}
