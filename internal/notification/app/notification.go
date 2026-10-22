package app

import (
	"encoding/json"
	"errors"
	"github.com/your-org/go-minimall/internal/notification/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Record(c *gin.Context) {
	var req struct {
		EventID   string          `json:"event_id"`
		EventType string          `json:"event_type"`
		UserID    string          `json:"user_id"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.EventID == "" || len(req.UserID) != 26 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	row, err := h.Repo.Record(req.EventID, req.EventType, req.UserID, req.Payload)
	if errors.Is(err, domain.ErrTemplateNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 19001, "message": "template not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 19002, "message": "failed to send"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": row})
}

func (h *Handler) ListMine(c *gin.Context) {
	rows, err := h.Repo.ListMine(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

func (h *Handler) MarkRead(c *gin.Context) {
	err := h.Repo.MarkRead(c.Param("id"), c.GetString("user_id"))
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": 19003, "message": "notification not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to mark read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}
