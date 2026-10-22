package app

import (
	"time"

	"github.com/your-org/go-minimall/internal/user/port"
)

type Handler struct {
	Repo       port.Repository
	JWTSecret  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewHandler(repo port.Repository) *Handler {
	return &Handler{
		Repo:       repo,
		JWTSecret:  []byte("dev-only-change-me"),
		AccessTTL:  30 * time.Second,
		RefreshTTL: 60 * time.Hour,
	}
}
