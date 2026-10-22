package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/product/port"
)

type Handler struct {
	DB   *sqlx.DB
	Repo port.Repository
}

func NewHandler(db *sqlx.DB, repo port.Repository) *Handler {
	return &Handler{DB: db, Repo: repo}
}
