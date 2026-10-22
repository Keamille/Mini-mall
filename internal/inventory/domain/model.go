package domain

import (
	"errors"
)

type ReserveItem struct {
	SkuID    string `db:"sku_id"`
	Quantity int    `db:"quantity"`
}

type StockRow struct {
	SkuID     string `db:"sku_id" json:"sku_id"`
	Available int    `db:"available" json:"available"`
	Reserved  int    `db:"reserved" json:"reserved"`
	Version   int    `db:"version" json:"version"`
}

var ErrInsufficientStock = errors.New("insufficient stock")

var ErrReservationNotFound = errors.New("reservation not found")

var ErrReservationCommitted = errors.New("reservation already committed")
