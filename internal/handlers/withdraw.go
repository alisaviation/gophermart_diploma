package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rtmelsov/GopherMart/internal/models"
	"github.com/theplant/luhn"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *Handler) PostBalanceWithdraw(c *gin.Context) {
	var req models.Withdrawal
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, "неверный номер заказа")
		return
	}
	num, err := strconv.Atoi(req.Number)
	if err != nil || luhn.Valid(num) {
		h.conf.GetLogger().Error("error", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, "неверный номер заказа")
		return
	}

	id := c.GetUint("userId")
	DBReq := models.DBWithdrawal{
		OrderNumber: req.Number,
		Sum:         req.Sum,
		UserID:      id,
	}

	localErr := h.serv.PostOrderWithDraw(&DBReq)
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

	code := http.StatusOK
	if len(*list) == 0 {
		code = http.StatusNoContent
	}

	c.Header("Content-Type", "application/json")
	c.JSON(code, *list)
}
