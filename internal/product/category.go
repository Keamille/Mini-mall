package product

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createCategoryRequest struct {
	ParentID *string `json:"parent_id"`
	Name     string  `json:"name" binding:"required"`
	Sort     int     `json:"sort"`
}

func (h *Handler) ListCategories(c *gin.Context) {
	rows, err := h.listCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rows})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}

	row := categoryRow{
		ID:     NewID(),
		Name:   req.Name,
		Sort:   req.Sort,
		Status: 1,
	}
	if req.ParentID != nil && *req.ParentID != "" {
		row.ParentID = sql.NullString{String: *req.ParentID, Valid: true}
	}

	if err := h.createCategory(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": row})
}
