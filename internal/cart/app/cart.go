package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type upsertCartRequest struct {
	SkuID    string `json:"sku_id" binding:"required"`
	Quantity int    `json:"quantity"`
}

func (h *Handler) ListByUser(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.Repo.ListByuser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10007,
			"message": "failed to list cart items",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    rows,
	})
}

func (h *Handler) UpsertItem(c *gin.Context) {
	userID := c.GetString("user_id")
	var req upsertCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "invalid request",
		})
		return
	}
	if req.Quantity <= 0 {
		if err := h.Repo.DeleteItem(userID, req.SkuID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    10007,
				"message": "failed to delete cart item",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "item deleted",
			"data":    nil,
		})
		return
	}
	if err := h.Repo.UpsertItem(userID, req.SkuID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10007,
			"message": "failed to upsert cart item",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user_id":  userID,
			"sku_id":   req.SkuID,
			"quantity": req.Quantity,
		},
	})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	skuID := c.Param("sku_id")
	userID := c.GetString("user_id")
	if err := h.Repo.DeleteItem(userID, skuID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10007,
			"message": "failed to delete cart item",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "item deleted",
		"data":    nil,
	})
}
