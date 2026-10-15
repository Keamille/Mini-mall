package product

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type createProductRequest struct {
	CategoryID  *string `json:"category_id"`
	Title       string  `json:"title" binding:"required"`
	Subtitle    string  `json:"subtitle"`
	Description *string `json:"description"`
	Status      *int8   `json:"status"` // 0草稿 1上架 2下架；默认 0
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}

	row := productRow{
		ID:       NewID(),
		Title:    req.Title,
		Subtitle: req.Subtitle,
		Status:   0,
	}
	if req.Status != nil {
		row.Status = *req.Status
	}
	if req.CategoryID != nil && *req.CategoryID != "" {
		row.CategoryID = sql.NullString{String: *req.CategoryID, Valid: true}
	}
	if req.Description != nil {
		row.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	if err := h.createProduct(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": row})
}

func (h *Handler) ListProducts(c *gin.Context) {
	onlyOnSale := c.Query("all") != "1"
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	products, err := h.listProducts(onlyOnSale, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": products})
}

func (h *Handler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	p, err := h.findProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 12001, "message": "product not found"})
		return
	}
	skus, err := h.listSkusByProductID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list skus"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"product": p, "skus": skus}})
}
