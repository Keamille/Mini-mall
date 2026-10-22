package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/order/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	auth := AuthRequired()
	r.POST("/api/v1/orders", auth, h.CreateOrder)
	r.POST("/api/v1/orders/:id/cancel", auth, h.CancelOrder)
	r.POST("/api/v1/orders/:id/paid", h.MarkPaid)
}
