package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/payment/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	auth := AuthRequired()
	r.POST("/api/v1/payments", auth, h.CreatePayment)
	r.POST("/api/v1/payments/:id/mock-confirm", auth, h.MockConfirm)
}
