package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/cart/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	auth := AuthRequired()
	r.GET("/api/v1/cart", auth, h.ListByUser)
	r.POST("/api/v1/cart", auth, h.UpsertItem)
	r.DELETE("/api/v1/cart/:sku_id", auth, h.DeleteItem)
}
