package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/notification/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	auth := AuthRequired()
	r.POST("/api/v1/notifications/events", h.Record)
	r.GET("/api/v1/notifications/mine", auth, h.ListMine)
	r.POST("/api/v1/notifications/:id/read", auth, h.MarkRead)
}
