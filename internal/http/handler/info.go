package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) GetInfo(c *gin.Context) {
	userID := c.GetInt("user_id")

	info, err := h.storage.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}
