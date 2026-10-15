package user

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// Handler 聚合 user 服务的 HTTP 处理依赖。
type Handler struct {
	DB         *sqlx.DB
	JWTSecret  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		DB:         db,
		JWTSecret:  []byte("dev-only-change-me"),
		AccessTTL:  30 * time.Second,
		RefreshTTL: 60 * time.Hour,
	}
}
