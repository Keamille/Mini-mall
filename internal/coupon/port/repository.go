package port

import (
	"github.com/your-org/go-minimall/internal/coupon/domain"
)

type Repository interface {
	ListTemplates() ([]domain.CouponRow, error)
	CreateTemplate(template domain.Coupon) error
	ClaimTx(userID, templateID string) (domain.UserCouponRow, error)
	ListMine(userID string) ([]domain.MineRow, error)
	QuoteCoupon(userID, userCouponID string, itemAmount int64) (int64, error)
	LockTx(userID, userCouponID, orderID string, itemAmount int64) (domain.CouponEffect, error)
	UnlockTx(userID, userCouponID, orderID string) error
	ConfirmTx(userID, userCouponID, orderID string) error
}
