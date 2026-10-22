package app

import (
	"errors"
	"github.com/your-org/go-minimall/internal/inventory/domain"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type reserveRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Items   []struct {
		SkuID    string `json:"sku_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	} `json:"items" binding:"required"`
}

func (h *Handler) Reserve(c *gin.Context) {
	var req reserveRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	items := make([]domain.ReserveItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "quantity must be > 0"})
			return
		}
		items = append(items, domain.ReserveItem{SkuID: it.SkuID, Quantity: it.Quantity})
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	id, err := h.Repo.ReserveTx(req.OrderID, expiresAt, items)
	if errors.Is(err, domain.ErrInsufficientStock) {
		c.JSON(http.StatusConflict, gin.H{"code": 13001, "message": "insufficient stock"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to reserve"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0, "message": "success",
		"data": gin.H{"reservation_id": id, "order_id": req.OrderID},
	})
}

func (h *Handler) Release(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	err := h.Repo.ReleaseTx(req.OrderID)
	if errors.Is(err, domain.ErrReservationNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": 13002, "message": "reservation not found"})
		return
	}
	if errors.Is(err, domain.ErrReservationCommitted) {
		c.JSON(http.StatusConflict, gin.H{"code": 13003, "message": "reservation already committed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to release"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *Handler) Commit(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	err := h.Repo.CommitTx(req.OrderID)
	if errors.Is(err, domain.ErrReservationNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": 13002, "message": "reservation not found"})
		return
	}
	if errors.Is(err, domain.ErrReservationCommitted) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 13003, "message": "reservation already committed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to commit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}
