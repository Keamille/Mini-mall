package product

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createSkuRequest struct {
	SkuCode   string          `json:"sku_code" binding:"required"`
	Attrs     json.RawMessage `json:"attrs"`
	PriceCent int64           `json:"price_cent" binding:"required"`
	Status    *int8           `json:"status"`
}

func (h *Handler) CreateSku(c *gin.Context) {
	productID := c.Param("id")
	if _, err := h.findProductByID(productID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 12001, "message": "product not found"})
		return
	}

	var req createSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	if req.PriceCent < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "price_cent must be >= 0"})
		return
	}

	row := skuRow{
		ID:        NewID(),
		ProductID: productID,
		SkuCode:   req.SkuCode,
		AttrsJSON: req.Attrs,
		PriceCent: req.PriceCent,
		Status:    1,
	}
	if req.Status != nil {
		row.Status = *req.Status
	}

	if err := h.createSku(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create sku"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": row})
}
