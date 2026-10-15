package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	u, err := h.findUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 10004, "message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       u.ID,
		"email":    u.Email.String,
		"phone":    u.Phone.String,
		"nickname": u.Nickname,
		"status":   u.Status,
	})
}
