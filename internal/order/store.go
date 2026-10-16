package order

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type itemRow struct {
	SkuID      string `json:"sku_id"`
	ProductID  string `json:"product_id"`
	Title      string `json:"title"`
	PriceCent  int64  `json:"price_cent"`
	Quantity   int    `json:"quantity"`
	AmountCent int64  `json:"amount_cent"`
}

type orderRow struct {
	ID             string         `db:"id" json:"id"`
	OrderNo        string         `db:"order_no" json:"order_no"`
	UserID         string         `db:"user_id" json:"user_id"`
	Status         int8           `db:"status" json:"status"`
	PayAmountCent  int64          `db:"pay_amount_cent" json:"pay_amount_cent"`
	ItemAmountCent int64          `db:"item_amount_cent" json:"item_amount_cent"`
	ReservationID  sql.NullString `db:"reservation_id" json:"reservation_id"`
	IdempotencyKey string         `db:"idempotency_key" json:"idempotency_key"`
}

func (h *Handler) withTx(fn func(tx *sqlx.Tx) error) error {
	tx, err := h.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (h *Handler) findByIdempotency(userID, key string) (orderRow, error) {
	var o orderRow
	err := h.DB.Get(&o, `
		SELECT id, order_no, user_id, status, pay_amount_cent, item_amount_cent, reservation_id, idempotency_key
		FROM orders WHERE user_id = ? AND idempotency_key = ?`, userID, key)
	return o, err
}

func (h *Handler) insertOrderTx(o orderRow, items []itemRow) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`
			INSERT INTO orders
			(id, order_no, user_id, status, pay_amount_cent, item_amount_cent, reservation_id, idempotency_key)
			VALUES (?, ?, ?, 10, ?, ?, ?, ?)`,
			o.ID, o.OrderNo, o.UserID, o.PayAmountCent, o.ItemAmountCent, o.ReservationID, o.IdempotencyKey,
		); err != nil {
			return err
		}
		for _, it := range items {
			if _, err := tx.Exec(`
				INSERT INTO order_items
				(id, order_id, sku_id, product_id, title, price_cent, quantity, amount_cent)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				NewID(), o.ID, it.SkuID, it.ProductID, it.Title, it.PriceCent, it.Quantity, it.AmountCent,
			); err != nil {
				return err
			}
		}
		return nil
	})
}
