package mysqlrepo

import (
	"fmt"

	"database/sql"

	"github.com/your-org/go-minimall/internal/coupon/domain"
	"github.com/your-org/go-minimall/internal/coupon/port"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) ListTemplates() ([]domain.CouponRow, error) {
	var templates []domain.CouponRow
	err := r.DB.Select(&templates, `
		SELECT id, name, type, threshold_cent, discount_cent, discount_rate,
        total_count, claimed_count, per_user_limit, valid_from, valid_to, status
		FROM coupon_templates
		WHERE status = 1 AND valid_from <= NOW(3) AND valid_to >= NOW(3)
		ORDER BY created_at DESC`)
	if templates == nil {
		templates = []domain.CouponRow{}
	}
	return templates, err
}

func (r *Repo) CreateTemplate(template domain.Coupon) error {
	if _, err := r.DB.Exec(`
		INSERT INTO coupon_templates (id, name, type, threshold_cent, discount_cent, discount_rate, total_count, claimed_count, per_user_limit, valid_from, valid_to, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, template.ID, template.Name, template.Type, template.ThresholdCent, template.DiscountCent, template.DiscountRate, template.TotalCount, template.ClaimedCount, template.PerUserLimit, template.ValidFrom, template.ValidTo, template.Status); err != nil {
		return err
	}
	return nil
}

func (r *Repo) ClaimTx(userID, templateID string) (domain.UserCouponRow, error) {
	var uc domain.UserCouponRow
	err := r.withTx(func(tx *sqlx.Tx) error {
		var tpl domain.CouponRow
		err := tx.Get(&tpl, `
			SELECT id, name, type, threshold_cent, discount_cent, discount_rate,
			       total_count, claimed_count, per_user_limit, valid_from, valid_to, status
			FROM coupon_templates
			WHERE id = ? AND status = 1
			  AND valid_from <= NOW(3) AND valid_to >= NOW(3)
			FOR UPDATE`, templateID)
		if err == sql.ErrNoRows {
			return domain.ErrTemplateNotClaimable
		}
		if err != nil {
			return err
		}
		if tpl.ClaimedCount >= tpl.TotalCount {
			return domain.ErrTemplateNotClaimable
		}
		var n int
		if err := tx.Get(&n, `
			SELECT COUNT(*) FROM user_coupons
			WHERE template_id = ? AND user_id = ?`, templateID, userID); err != nil {
			return err
		}
		if int64(n) >= tpl.PerUserLimit {
			return domain.ErrClaimLimit
		}
		uc = domain.UserCouponRow{
			ID:         domain.NewID(),
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
			return domain.ErrTemplateNotClaimable
		}
		return nil
	})
	return uc, err
}

func (r *Repo) ListMine(userID string) ([]domain.MineRow, error) {
	var coupons []domain.MineRow
	err := r.DB.Select(&coupons, `
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
		coupons = []domain.MineRow{}
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

func (r *Repo) loadCoupon(q sqlx.Queryer, userCouponID string, lock bool) (domain.CouponState, error) {
	query := couponStateSQL
	if lock {
		query += ` FOR UPDATE`
	}
	var row domain.CouponState
	err := sqlx.Get(q, &row, query, userCouponID)
	return row, err
}

func ownedCoupon(row domain.CouponState, err error, userID string) (domain.CouponState, error) {
	if err == sql.ErrNoRows {
		return domain.CouponState{}, domain.ErrCouponNotFound
	}
	if err != nil {
		return domain.CouponState{}, err
	}
	if row.UserID != userID {
		return domain.CouponState{}, domain.ErrCouponNotFound
	}
	return row, nil
}

func (r *Repo) QuoteCoupon(userID, userCouponID string, itemAmount int64) (int64, error) {
	row, err := r.loadCoupon(r.DB, userCouponID, false)
	row, err = ownedCoupon(row, err, userID)
	if err != nil {
		return 0, err
	}
	if err := usableCoupon(row, itemAmount); err != nil {
		return 0, err
	}
	return discountCent(row.DiscountCent, itemAmount), nil
}

func usableCoupon(row domain.CouponState, itemAmount int64) error {
	if row.Expired == 1 {
		return domain.ErrCouponExpired
	}
	if row.Status != 1 {
		return domain.ErrCouponStatus
	}
	if itemAmount < row.ThresholdCent {
		return domain.ErrCouponThreshold
	}
	return nil
}

func (r *Repo) LockTx(userID, userCouponID, orderID string, itemAmount int64) (domain.CouponEffect, error) {
	var out domain.CouponEffect
	err := r.withTx(func(tx *sqlx.Tx) error {
		row, err := r.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		discount := discountCent(row.DiscountCent, itemAmount)
		if row.Status == 2 && row.LockedOrderID == orderID {
			if row.Expired == 1 {
				return domain.ErrCouponExpired
			}
			if itemAmount < row.ThresholdCent {
				return domain.ErrCouponThreshold
			}
			out = domain.CouponEffect{UserCouponID: row.ID, DiscountCent: discount, Status: 2}
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
			return domain.ErrCouponStatus
		}
		out = domain.CouponEffect{UserCouponID: row.ID, DiscountCent: discount, Status: 2}
		return nil
	})
	return out, err
}

func (r *Repo) UnlockTx(userID, userCouponID, orderID string) error {
	return r.withTx(func(tx *sqlx.Tx) error {
		row, err := r.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		if row.Status == 1 {
			return nil
		}
		if row.Status != 2 || row.LockedOrderID != orderID {
			return domain.ErrCouponStatus
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
			return domain.ErrCouponStatus
		}
		return nil
	})
}

func (r *Repo) ConfirmTx(userID, userCouponID, orderID string) error {
	return r.withTx(func(tx *sqlx.Tx) error {
		row, err := r.loadCoupon(tx, userCouponID, true)
		row, err = ownedCoupon(row, err, userID)
		if err != nil {
			return err
		}
		if row.Status == 3 && row.LockedOrderID == orderID {
			return nil
		}
		if row.Status != 2 || row.LockedOrderID != orderID {
			return domain.ErrCouponStatus
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
			return domain.ErrCouponStatus
		}
		return nil
	})
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
