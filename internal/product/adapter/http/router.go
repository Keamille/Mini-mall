package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/product/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	r.GET("/api/v1/categories", h.ListCategories)
	r.POST("/api/v1/categories", h.CreateCategory)
	r.GET("/api/v1/products", h.ListProducts)
	r.POST("/api/v1/products", h.CreateProduct)
	r.GET("/api/v1/products/:id", h.GetProduct)
	r.POST("/api/v1/products/:id/skus", h.CreateSku)
}
