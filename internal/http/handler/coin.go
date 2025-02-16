package handler

import (
	"avito-tech-winter-2025/internal/storage"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func (h *Handler) SendCoin(c *gin.Context) {
	var req SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	fromUserID := c.GetInt("user_id")
	err := h.storage.TransferCoins(c.Request.Context(), fromUserID, req.ToUser, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, storage.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Status(http.StatusOK)
}
