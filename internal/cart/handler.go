package cart

import "github.com/jmoiron/sqlx"

type Handler struct {
	DB *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{DB: db}
}
