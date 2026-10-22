package app

import "github.com/your-org/go-minimall/internal/inventory/port"

type Handler struct {
	Repo port.Repository
}

func NewHandler(repo port.Repository) *Handler {
	return &Handler{Repo: repo}
}
