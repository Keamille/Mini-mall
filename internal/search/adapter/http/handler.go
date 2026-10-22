package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/search/app"
	"github.com/your-org/go-minimall/internal/search/domain"
)

func Register(r *gin.Engine, svc *app.Service) {
	r.GET("/api/v1/search/products", func(c *gin.Context) { searchProducts(c, svc) })
	r.POST("/api/v1/search/products", func(c *gin.Context) { upsert(c, svc) })
}

func upsert(c *gin.Context, svc *app.Service) {
	var doc domain.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid search document"})
		return
	}
	if err := svc.Upsert(c.Request.Context(), doc); err != nil {
		if errors.Is(err, domain.ErrBadDocument) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid search document"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 18001, "message": "search unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func searchProducts(c *gin.Context, svc *app.Service) {
	q := c.Query("q")
	if len(q) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid query"})
		return
	}
	categoryID := c.Query("category_id")
	if categoryID != "" && len(categoryID) != 26 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid category_id"})
		return
	}
	priceMin, ok := optionalCent(c, "price_min")
	if !ok {
		return
	}
	priceMax, ok := optionalCent(c, "price_max")
	if !ok {
		return
	}
	if priceMin != nil && priceMax != nil && *priceMin > *priceMax {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid price range"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit <= 0 || limit > 100 || page <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid page"})
		return
	}
	rows, err := svc.Search(c.Request.Context(), domain.Query{
		Text:       q,
		CategoryID: categoryID,
		PriceMin:   priceMin,
		PriceMax:   priceMax,
		Sort:       c.Query("sort"),
		Limit:      limit,
		Offset:     (page - 1) * limit,
	})
	if errors.Is(err, domain.ErrBadQuery) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid sort"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 18001, "message": "search unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"items": rows, "page": page, "page_size": limit}})
}

func optionalCent(c *gin.Context, key string) (*int64, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 18002, "message": "invalid " + key})
		return nil, false
	}
	return &n, true
}
