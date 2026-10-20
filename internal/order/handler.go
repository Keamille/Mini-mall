package order

import (
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	DB *sqlx.DB
	js jetstream.JetStream
}

func NewHandler(db *sqlx.DB, js jetstream.JetStream) *Handler {
	return &Handler{DB: db, js: js}
}
