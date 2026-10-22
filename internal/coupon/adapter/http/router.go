package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/coupon/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	auth := AuthRequired()
	r.GET("/api/v1/coupons/templates", h.ListTemplates)
	r.POST("/api/v1/coupons/templates", auth, h.CreateTemplate)
	r.POST("/api/v1/coupons/templates/:id/claim", auth, h.Claim)
	r.GET("/api/v1/coupons/mine", auth, h.ListMine)
	r.POST("/api/v1/coupons/quote", auth, h.Quote)
	r.POST("/api/v1/coupons/lock", h.Lock)
	r.POST("/api/v1/coupons/unlock", h.Unlock)
	r.POST("/api/v1/coupons/confirm", h.Confirm)
}
