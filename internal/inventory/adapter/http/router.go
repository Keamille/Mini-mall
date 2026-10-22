package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/inventory/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	r.POST("/api/v1/inventories", h.InitStock)
	r.GET("/api/v1/inventories/:sku_id", h.GetStock)
	r.POST("/api/v1/inventories/reserve", h.Reserve)
	r.POST("/api/v1/inventories/release", h.Release)
	r.POST("/api/v1/inventories/commit", h.Commit)
}
