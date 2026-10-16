package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type initStockRequest struct {
	SkuID     string `json:"sku_id" binding:"required"`
	Available int    `json:"available" binding:"required"`
}

func (h *Handler) InitStock(c *gin.Context) {
	var req initStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	if req.Available < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "available must be >= 0"})
		return
	}
	if err := h.initStock(req.SkuID, req.Available); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to init stock"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 0, "message": "success",
		"data": stockRow{SkuID: req.SkuID, Available: req.Available},
	})
}

func (h *Handler) GetStock(c *gin.Context) {
	s, err := h.findStock(c.Param("sku_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 13002, "message": "stock not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": s})
}
