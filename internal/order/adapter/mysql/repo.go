package mysqlrepo

import (
	"errors"
	"fmt"
	"log"

	"github.com/your-org/go-minimall/internal/order/domain"
	"github.com/your-org/go-minimall/internal/order/port"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

// EnsureSchema adds pay_expire_at on existing DBs that predate the column.
func EnsureSchema(db *sqlx.DB) {
	if _, err := db.Exec(`ALTER TABLE orders ADD COLUMN pay_expire_at DATETIME(3) NULL`); err != nil {
		var me *mysql.MySQLError
		if !errors.As(err, &me) || me.Number != 1060 {
			log.Printf("order ensure pay_expire_at: %v", err)
		}
	}
	if _, err := db.Exec(`ALTER TABLE orders ADD KEY idx_orders_pay_expire (status, pay_expire_at)`); err != nil {
		var me *mysql.MySQLError
		if !errors.As(err, &me) || me.Number != 1061 {
			log.Printf("order ensure idx_orders_pay_expire: %v", err)
		}
	}
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

func (r *Repo) FindByIdempotency(userID, key string) (domain.OrderRow, error) {
	var o domain.OrderRow
	err := r.DB.Get(&o, `
		SELECT id, order_no, user_id, status, pay_amount_cent, item_amount_cent,
		       discount_cent, coupon_id, reservation_id, idempotency_key, pay_expire_at
		FROM orders WHERE user_id = ? AND idempotency_key = ?`, userID, key)
	return o, err
}

func (r *Repo) InsertOrderTx(o domain.OrderRow, items []domain.ItemRow) error {
	return r.withTx(func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`
			INSERT INTO orders
			(id, order_no, user_id, status, pay_amount_cent, item_amount_cent, discount_cent, coupon_id, reservation_id, idempotency_key, pay_expire_at)
			VALUES (?, ?, ?, 10, ?, ?, ?, ?, ?, ?, ?)`,
			o.ID, o.OrderNo, o.UserID, o.PayAmountCent, o.ItemAmountCent, o.DiscountCent, o.CouponID, o.ReservationID, o.IdempotencyKey, o.PayExpireAt,
		); err != nil {
			return err
		}
		for _, it := range items {
			if _, err := tx.Exec(`
				INSERT INTO order_items
				(id, order_id, sku_id, product_id, title, price_cent, quantity, amount_cent)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				domain.NewID(), o.ID, it.SkuID, it.ProductID, it.Title, it.PriceCent, it.Quantity, it.AmountCent,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repo) FindOrderByID(id string, userID string) (domain.OrderRow, error) {
	var o domain.OrderRow
	err := r.DB.Get(&o, `
		SELECT id, order_no, user_id, status, pay_amount_cent, item_amount_cent,
		       discount_cent, coupon_id, reservation_id, idempotency_key, pay_expire_at
		FROM orders WHERE id = ? AND user_id = ?`, id, userID)
	return o, err
}

func (r *Repo) MarkCancelled(id string, userID string) (bool, error) {
	res, err := r.DB.Exec(`
		UPDATE orders
		SET status = 30, cancelled_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND user_id = ? AND status = 10`, id, userID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *Repo) FindOrder(id string) (domain.OrderRow, error) {
	var o domain.OrderRow
	err := r.DB.Get(&o, `
		SELECT id, order_no, user_id, status, pay_amount_cent, item_amount_cent,
		       discount_cent, coupon_id, reservation_id, idempotency_key, pay_expire_at
		FROM orders WHERE id = ?`, id)
	return o, err
}

func (r *Repo) MarkPaid(id string) (bool, error) {
	res, err := r.DB.Exec(`
		UPDATE orders SET status = 20, paid_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND status = 10`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (r *Repo) ListExpiredPending(limit int) ([]domain.OrderRow, error) {
	var rows []domain.OrderRow
	err := r.DB.Select(&rows, `
		SELECT id, order_no, user_id, status, pay_amount_cent, item_amount_cent,
		       discount_cent, coupon_id, reservation_id, idempotency_key, pay_expire_at
		FROM orders
		WHERE status = 10 AND pay_expire_at IS NOT NULL AND pay_expire_at <= UTC_TIMESTAMP(3)
		ORDER BY pay_expire_at ASC
		LIMIT ?`, limit)
	return rows, err
}

func (r *Repo) MarkClosed(id string) (bool, error) {
	res, err := r.DB.Exec(`
		UPDATE orders SET status = 40
		WHERE id = ? AND status = 10`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
