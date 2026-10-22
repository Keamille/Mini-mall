package domain

import (
	"database/sql"
	"time"
)

type UserRow struct {
	ID           string         `db:"id"`
	Email        sql.NullString `db:"email"`
	Phone        sql.NullString `db:"phone"`
	PasswordHash string         `db:"password_hash"`
	Nickname     string         `db:"nickname"`
	Status       int8           `db:"status"`
}

type RefreshTokenRow struct {
	ID        string       `db:"id"`
	UserID    string       `db:"user_id"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
}

type AddressRow struct {
	ID           string `db:"id"`
	UserID       string `db:"user_id"`
	ReceiverName string `db:"receiver_name"`
	Phone        string `db:"phone"`
	Province     string `db:"province"`
	City         string `db:"city"`
	District     string `db:"district"`
	Detail       string `db:"detail"`
	IsDefault    int8   `db:"is_default"`
}
