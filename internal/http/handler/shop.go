package handler

import (
	"avito-tech-winter-2025/internal/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) BuyItem(c *gin.Context) {
	item := c.Param("item")
	userID := c.GetInt("user_id")

	err := h.storage.BuyItem(c.Request.Context(), userID, item)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidItem):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item"})
		case errors.Is(err, storage.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Status(http.StatusOK)
}
