package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/user/app"
)

func Register(r *gin.Engine, h *app.Handler) {
	r.POST("/api/v1/auth/register", h.Register_user)
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/refresh", h.Refresh)
	auth := r.Group("/api/v1")
	auth.Use(h.AuthRequired())
	auth.GET("/users/me", h.Me)
	auth.GET("/users/me/addresses", h.ListAddress)
	auth.POST("/users/me/addresses", h.CreateAddress)
	auth.PUT("/users/me/addresses/:id", h.UpdateAddress)
	auth.DELETE("/users/me/addresses/:id", h.DeleteAddress)
}
