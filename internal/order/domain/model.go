package domain

import (
	"database/sql"
	"errors"
)

type ItemRow struct {
	SkuID      string `json:"sku_id"`
	ProductID  string `json:"product_id"`
	Title      string `json:"title"`
	PriceCent  int64  `json:"price_cent"`
	Quantity   int    `json:"quantity"`
	AmountCent int64  `json:"amount_cent"`
}

type OrderRow struct {
	ID             string         `db:"id" json:"id"`
	OrderNo        string         `db:"order_no" json:"order_no"`
	UserID         string         `db:"user_id" json:"user_id"`
	Status         int8           `db:"status" json:"status"`
	PayAmountCent  int64          `db:"pay_amount_cent" json:"pay_amount_cent"`
	ItemAmountCent int64          `db:"item_amount_cent" json:"item_amount_cent"`
	DiscountCent   int64          `db:"discount_cent" json:"discount_cent"`
	CouponID       sql.NullString `db:"coupon_id" json:"coupon_id"`
	ReservationID  sql.NullString `db:"reservation_id" json:"reservation_id"`
	IdempotencyKey string         `db:"idempotency_key" json:"idempotency_key"`
	PayExpireAt    sql.NullTime   `db:"pay_expire_at" json:"pay_expire_at"`
}

var ErrInsufficientStock = errors.New("insufficient stock")

var ErrReservationNotFound = errors.New("reservation not found")

var ErrReservationCommitted = errors.New("reservation already committed")
