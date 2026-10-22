package mysqlrepo

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/inventory/domain"
	"github.com/your-org/go-minimall/internal/inventory/port"
	"time"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) FindStock(skuID string) (domain.StockRow, error) {
	var s domain.StockRow
	err := r.DB.Get(&s, `
		SELECT sku_id, available, reserved, version
		FROM stocks WHERE sku_id = ?`, skuID)
	return s, err
}

func (r *Repo) InitStock(skuID string, available int) error {
	_, err := r.DB.Exec(`
		INSERT INTO stocks (sku_id, available, reserved, version)
		VALUES (?, ?, 0, 0)`,
		skuID, available,
	)
	return err
}

func (r *Repo) withTx(fn func(tx *sqlx.Tx) error) error {
	tx, err := r.DB.Beginx()
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

func (r *Repo) ReserveTx(orderID string, expiresAt time.Time, items []domain.ReserveItem) (string, error) {
	var reservationID string
	err := r.withTx(func(tx *sqlx.Tx) error {
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
				return domain.ErrInsufficientStock
			}
		}
		reservationID = domain.NewID()
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
				domain.NewID(), reservationID, it.SkuID, it.Quantity,
			); err != nil {
				return err
			}
		}
		return nil
	})
	return reservationID, err
}

func (r *Repo) ReleaseTx(orderID string) error {
	return r.withTx(func(tx *sqlx.Tx) error {
		var row struct {
			ID     string `db:"id"`
			Status int8   `db:"status"`
		}
		err := tx.Get(&row, `SELECT id, status FROM stock_reservations WHERE order_id = ?`, orderID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrReservationNotFound
		}
		if err != nil {
			return err
		}
		if row.Status == 3 {
			return nil
		}
		if row.Status != 1 {
			return domain.ErrReservationCommitted
		}

		var items []domain.ReserveItem
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

func (r *Repo) CommitTx(orderID string) error {
	return r.withTx(func(tx *sqlx.Tx) error {
		var row struct {
			ID     string `db:"id"`
			Status int8   `db:"status"`
		}
		err := tx.Get(&row, `SELECT id, status FROM stock_reservations WHERE order_id = ?`, orderID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrReservationNotFound
		}
		if err != nil {
			return err
		}
		if row.Status == 2 {
			return nil
		}
		if row.Status != 1 {
			return domain.ErrReservationCommitted
		}

		var items []domain.ReserveItem
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
