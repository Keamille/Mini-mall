package coupon

import (
	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	ErrTemplateNotClaimable = errors.New("template not claimable")
	ErrClaimLimit           = errors.New("claim limit reached")
	ErrCouponNotFound       = errors.New("coupon not found")
	ErrCouponStatus         = errors.New("coupon status invalid")
	ErrCouponThreshold      = errors.New("threshold not met")
	ErrCouponExpired        = errors.New("coupon expired")
)

type couponRow struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Type          int64     `db:"type" json:"type"`
	ThresholdCent int64     `db:"threshold_cent" json:"threshold_cent"`
	DiscountCent  int64     `db:"discount_cent" json:"discount_cent"`
	DiscountRate  int64     `db:"discount_rate" json:"discount_rate"`
	TotalCount    int64     `db:"total_count" json:"total_count"`
	ClaimedCount  int64     `db:"claimed_count" json:"claimed_count"`
	PerUserLimit  int64     `db:"per_user_limit" json:"per_user_limit"`
	ValidFrom     time.Time `db:"valid_from" json:"valid_from"`
	ValidTo       time.Time `db:"valid_to" json:"valid_to"`
	Status        int       `db:"status" json:"status"`
}
type userCouponRow struct {
	ID         string `db:"id" json:"id"`
	TemplateID string `db:"template_id" json:"template_id"`
	UserID     string `db:"user_id" json:"user_id"`
	Status     int    `db:"status" json:"status"`
}

type mineRow struct {
	ID            string    `db:"id" json:"id"`
	TemplateID    string    `db:"template_id" json:"template_id"`
	Status        int       `db:"status" json:"status"`
	LockedOrderID string    `db:"locked_order_id" json:"locked_order_id"`
	Name          string    `db:"name" json:"name"`
	ThresholdCent int64     `db:"threshold_cent" json:"threshold_cent"`
	DiscountCent  int64     `db:"discount_cent" json:"discount_cent"`
	ValidTo       time.Time `db:"valid_to" json:"valid_to"`
}

type couponState struct {
	ID            string `db:"id"`
	UserID        string `db:"user_id"`
	Status        int    `db:"status"`
	LockedOrderID string `db:"locked_order_id"`
	ThresholdCent int64  `db:"threshold_cent"`
	DiscountCent  int64  `db:"discount_cent"`
	Expired       int    `db:"expired"`
}

type couponEffect struct {
	UserCouponID string `json:"user_coupon_id"`
	DiscountCent int64  `json:"discount_cent"`
	Status       int    `json:"status"`
}

func (h *Handler) listTemplates() ([]couponRow, error) {
	var templates []couponRow
	err := h.DB.Select(&templates, `
		SELECT id, name, type, threshold_cent, discount_cent, discount_rate,
        total_count, claimed_count, per_user_limit, valid_from, valid_to, status
		FROM coupon_templates
		WHERE status = 1 AND valid_from <= NOW(3) AND valid_to >= NOW(3)
		ORDER BY created_at DESC`)
	if templates == nil {
		templates = []couponRow{}
	}
	return templates, err
}

func (h *Handler) createTemplate(template Coupon) error {
	if _, err := h.DB.Exec(`
		INSERT INTO coupon_templates (id, name, type, threshold_cent, discount_cent, discount_rate, total_count, claimed_count, per_user_limit, valid_from, valid_to, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, template.ID, template.Name, template.Type, template.ThresholdCent, template.DiscountCent, template.DiscountRate, template.TotalCount, template.ClaimedCount, template.PerUserLimit, template.ValidFrom, template.ValidTo, template.Status); err != nil {
		return err
	}
	return nil
}

func (h *Handler) claimTx(userID, templateID string) (userCouponRow, error) {
	var uc userCouponRow
	err := h.withTx(func(tx *sqlx.Tx) error {
		var tpl couponRow
		err := tx.Get(&tpl, `
			SELECT id, name, type, threshold_cent, discount_cent, discount_rate,
			       total_count, claimed_count, per_user_limit, valid_from, valid_to, status
			FROM coupon_templates
			WHERE id = ? AND status = 1
			  AND valid_from <= NOW(3) AND valid_to >= NOW(3)
			FOR UPDATE`, templateID)
		if err == sql.ErrNoRows {
			return ErrTemplateNotClaimable
		}
		if err != nil {
			return err
		}
		if tpl.ClaimedCount >= tpl.TotalCount {
			return ErrTemplateNotClaimable
		}
		var n int
		if err := tx.Get(&n, `
			SELECT COUNT(*) FROM user_coupons
			WHERE template_id = ? AND user_id = ?`, templateID, userID); err != nil {
			return err
		}
		if int64(n) >= tpl.PerUserLimit {
			return ErrClaimLimit
		}
		uc = userCouponRow{
			ID:         NewID(),
			TemplateID: templateID,
			UserID:     userID,
			Status:     1,
		}
		if _, err := tx.Exec(`
			INSERT INTO user_coupons (id, template_id, user_id, status)
			VALUES (?, ?, ?, 1)`, uc.ID, uc.TemplateID, uc.UserID); err != nil {
			return err
		}
		res, err := tx.Exec(`
			UPDATE coupon_templates
			SET claimed_count = claimed_count + 1
			WHERE id = ? AND claimed_count < total_count`, templateID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrTemplateNotClaimable
		}
		return nil
	})
	return uc, err
}

func (h *Handler) listMine(userID string) ([]mineRow, error) {
	var coupons []mineRow
	err := h.DB.Select(&coupons, `
		SELECT uc.id, uc.template_id, uc.status,
       COALESCE(uc.locked_order_id, '') AS locked_order_id,
       t.name, t.threshold_cent, t.discount_cent, t.valid_to
	FROM user_coupons uc
	JOIN coupon_templates t ON t.id = uc.template_id
	WHERE uc.user_id = ?
 	ORDER BY uc.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if coupons == nil {
		coupons = []mineRow{}
	}
	return coupons, nil
}

func discountCent(templateDiscount, itemAmount int64) int64 {
	if templateDiscount > itemAmount {
		return itemAmount
	}
	return templateDiscount
}

const couponStateSQL = `
	SELECT uc.id, uc.user_id, uc.status,
	       COALESCE(uc.locked_order_id, '') AS locked_order_id,
	       t.threshold_cent, t.discount_cent,
	       IF(t.valid_to < NOW(3), 1, 0) AS expired
	FROM user_coupons uc
	JOIN coupon_templates t ON t.id = uc.template_id
	WHERE uc.id = ?`

func (h *Handler) loadCoupon(q sqlx.Queryer, userCouponID string, lock bool) (couponState, error) {
	query := couponStateSQL
	if lock {
		query += ` FOR UPDATE`
	}
	var row couponState
	err := sqlx.Get(q, &row, query, userCouponID)
	return row, err
}

func ownedCoupon(row couponState, err error, userID string) (couponState, error) {
	if err == sql.ErrNoRows {
		return couponState{}, ErrCouponNotFound
	}
	if err != nil {
		return couponState{}, err
	}
	if row.UserID != userID {
		return couponState{}, ErrCouponNotFound
	}
	return row, nil
}

func (h *Handler) quoteCoupon(userID, userCouponID string, itemAmount int64) (int64, error) {
	row, err := h.loadCoupon(h.DB, userCouponID, false)
	row, err = ownedCoupon(row, err, userID)
	if err != nil {
		return 0, err
	}
	if err := usableCoupon(row, itemAmount); err != nil {
		return 0, err
	}
	return discountCent(row.DiscountCent, itemAmount), nil
}

func usableCoupon(row couponState, itemAmount int64) error {
	if row.Expired == 1 {
		return ErrCouponExpired
	}
	if row.Status != 1 {
		return ErrCouponStatus
	}
	if itemAmount < row.ThresholdCent {
		return ErrCouponThreshold
	}
	return nil
}

func (h *Handler) lockTx(userID, userCouponID, orderID string, itemAmount int64) (couponEffect, error) {
	var out couponEffect
	err := h.withTx(func(tx *sqlx.Tx) error {
		row, err := h.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		discount := discountCent(row.DiscountCent, itemAmount)
		if row.Status == 2 && row.LockedOrderID == orderID {
			if row.Expired == 1 {
				return ErrCouponExpired
			}
			if itemAmount < row.ThresholdCent {
				return ErrCouponThreshold
			}
			out = couponEffect{UserCouponID: row.ID, DiscountCent: discount, Status: 2}
			return nil
		}
		if err := usableCoupon(row, itemAmount); err != nil {
			return err
		}
		res, err := tx.Exec(`
			UPDATE user_coupons
			SET status = 2, locked_order_id = ?
			WHERE id = ? AND user_id = ? AND status = 1`,
			orderID, userCouponID, userID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrCouponStatus
		}
		out = couponEffect{UserCouponID: row.ID, DiscountCent: discount, Status: 2}
		return nil
	})
	return out, err
}

func (h *Handler) unlockTx(userID, userCouponID, orderID string) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		row, err := h.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		if row.Status == 1 {
			return nil
		}
		if row.Status != 2 || row.LockedOrderID != orderID {
			return ErrCouponStatus
		}
		res, err := tx.Exec(`
			UPDATE user_coupons
			SET status = 1, locked_order_id = NULL
			WHERE id = ? AND user_id = ? AND status = 2 AND locked_order_id = ?`,
			userCouponID, userID, orderID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrCouponStatus
		}
		return nil
	})
}

func (h *Handler) confirmTx(userID, userCouponID, orderID string) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		row, err := h.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		if row.Status == 3 && row.LockedOrderID == orderID {
			return nil
		}
		if row.Status != 2 || row.LockedOrderID != orderID {
			return ErrCouponStatus
		}
		res, err := tx.Exec(`
			UPDATE user_coupons
			SET status = 3, used_at = UTC_TIMESTAMP(3)
			WHERE id = ? AND user_id = ? AND status = 2 AND locked_order_id = ?`,
			userCouponID, userID, orderID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrCouponStatus
		}
		return nil
	})
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
