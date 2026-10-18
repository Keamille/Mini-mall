package inventory

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"github.com/jmoiron/sqlx"
)

var ErrInsufficientStock = errors.New("insufficient stock")
var	ErrReservationNotFound  = errors.New("reservation not found")
var	ErrReservationCommitted = errors.New("reservation already committed")

type reserveItem struct {
	SkuID    string `db:"sku_id"`
	Quantity int    `db:"quantity"`
}

type stockRow struct {
	SkuID     string `db:"sku_id" json:"sku_id"`
	Available int    `db:"available" json:"available"`
	Reserved  int    `db:"reserved" json:"reserved"`
	Version   int    `db:"version" json:"version"`
}

func (h *Handler) findStock(skuID string) (stockRow, error) {
	var s stockRow
	err := h.DB.Get(&s, `
		SELECT sku_id, available, reserved, version
		FROM stocks WHERE sku_id = ?`, skuID)
	return s, err
}

func (h *Handler) initStock(skuID string, available int) error {
	_, err := h.DB.Exec(`
		INSERT INTO stocks (sku_id, available, reserved, version)
		VALUES (?, ?, 0, 0)`,
		skuID, available,
	)
	return err
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

func (h *Handler) reserveTx(orderID string, expiresAt time.Time, items []reserveItem) (string, error) {
	var reservationID string
	err := h.withTx(func(tx *sqlx.Tx) error {
		var existing string
		err := tx.Get(&existing, `SELECT id FROM stock_reservations WHERE order_id = ?`, orderID)
		if err == nil {
			reservationID = existing
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		for _, it := range items {
			var version int
			err := tx.Get(&version, `SELECT version FROM stocks WHERE sku_id = ?`, it.SkuID)
			if err != nil {
				return err
			}
			res, err := tx.Exec(`
				UPDATE stocks
				SET available = available - ?, reserved = reserved + ?, version = version + 1
				WHERE sku_id = ? AND version = ? AND available >= ?`,
				it.Quantity, it.Quantity, it.SkuID, version, it.Quantity,
			)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				return ErrInsufficientStock
			}
		}
		reservationID = NewID()
		if _, err := tx.Exec(`
			INSERT INTO stock_reservations (id, order_id, status, expires_at)
			VALUES (?, ?, 1, ?)`,
			reservationID, orderID, expiresAt,
		); err != nil {
			return err
		}
		for _, it := range items {
			if _, err := tx.Exec(`
				INSERT INTO stock_reservation_items (id, reservation_id, sku_id, quantity)
				VALUES (?, ?, ?, ?)`,
				NewID(), reservationID, it.SkuID, it.Quantity,
			); err != nil {
				return err
			}
		}
		return nil
	})
	return reservationID, err
}

func (h *Handler) releaseTx(orderID string) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		var row struct {
			ID     string `db:"id"`
			Status int8   `db:"status"`
		}
		err := tx.Get(&row, `SELECT id, status FROM stock_reservations WHERE order_id = ?`, orderID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReservationNotFound
		}
		if err != nil {
			return err
		}
		if row.Status == 3 {
			return nil
		}
		if row.Status != 1 {
			return ErrReservationCommitted
		}

		var items []reserveItem
		if err := tx.Select(&items, `
			SELECT sku_id, quantity
			FROM stock_reservation_items WHERE reservation_id = ?`, row.ID); err != nil {
			return err
		}

		for _, it := range items {
			var version int
			if err := tx.Get(&version, `SELECT version FROM stocks WHERE sku_id = ?`, it.SkuID); err != nil {
				return err
			}
			res, err := tx.Exec(`
				UPDATE stocks
				SET available = available + ?, reserved = reserved - ?, version = version + 1
				WHERE sku_id = ? AND version = ? AND reserved >= ?`,
				it.Quantity, it.Quantity, it.SkuID, version, it.Quantity,
			)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				return fmt.Errorf("release stock failed for sku %s", it.SkuID)
			}
		}

		_, err = tx.Exec(`UPDATE stock_reservations SET status = 3 WHERE id = ? AND status = 1`, row.ID)
		return err
	})
}

func (h *Handler) commitTx(orderID string) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		var row struct {
			ID     string `db:"id"`
			Status int8   `db:"status"`
		}
		err := tx.Get(&row, `SELECT id, status FROM stock_reservations WHERE order_id = ?`, orderID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReservationNotFound
		}
		if err != nil {
			return err
		}
		if row.Status == 2 {
			return nil
		}
		if row.Status != 1 {
			return ErrReservationCommitted
		}

		var items []reserveItem
		if err := tx.Select(&items, `
			SELECT sku_id, quantity
			FROM stock_reservation_items WHERE reservation_id = ?`, row.ID); err != nil {
			return err
		}
		for _, it := range items {
			var version int
			if err := tx.Get(&version, `SELECT version FROM stocks WHERE sku_id = ?`, it.SkuID); err != nil {
				return err
			}
			res, err := tx.Exec(`
				UPDATE stocks
				SET reserved = reserved - ?, version = version + 1
				WHERE sku_id = ? AND version = ? AND reserved >= ?`,
				it.Quantity, it.SkuID, version, it.Quantity,
			)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				return fmt.Errorf("commit stock failed for sku %s", it.SkuID)
			}
		}
		_, err = tx.Exec(`UPDATE stock_reservations SET status = 2 WHERE id = ? AND status = 1`, row.ID)
		return err
	})
}
