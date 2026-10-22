package app

import (
	"errors"
	"github.com/your-org/go-minimall/internal/coupon/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AmountRequest struct {
	UserCouponID   string `json:"user_coupon_id"`
	UserID         string `json:"user_id"`
	OrderID        string `json:"order_id"`
	ItemAmountCent int64  `json:"item_amount_cent"`
}

func (h *Handler) Quote(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		UserCouponID   string `json:"user_coupon_id"`
		ItemAmountCent int64  `json:"item_amount_cent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserCouponID == "" || req.ItemAmountCent < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	discount, err := h.Repo.QuoteCoupon(userID, req.UserCouponID, req.ItemAmountCent)
	if writeCouponErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"discount_cent": discount}})
}

func (h *Handler) Lock(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil || !ValidLockRequest(req) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	out, err := h.Repo.LockTx(req.UserID, req.UserCouponID, req.OrderID, req.ItemAmountCent)
	if writeCouponErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": out})
}

func (h *Handler) Unlock(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserCouponID == "" || req.UserID == "" || len(req.OrderID) != 26 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	if writeCouponErr(c, h.Repo.UnlockTx(req.UserID, req.UserCouponID, req.OrderID)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *Handler) Confirm(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserCouponID == "" || req.UserID == "" || len(req.OrderID) != 26 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	if writeCouponErr(c, h.Repo.ConfirmTx(req.UserID, req.UserCouponID, req.OrderID)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func ValidLockRequest(req AmountRequest) bool {
	return req.UserCouponID != "" && req.UserID != "" && len(req.OrderID) == 26 && req.ItemAmountCent >= 0
}

func writeCouponErr(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, domain.ErrTemplateNotClaimable):
		c.JSON(http.StatusNotFound, gin.H{"code": 17001, "message": "template not claimable"})
	case errors.Is(err, domain.ErrClaimLimit):
		c.JSON(http.StatusConflict, gin.H{"code": 17002, "message": "claim limit reached"})
	case errors.Is(err, domain.ErrCouponNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": 17003, "message": "coupon not found"})
	case errors.Is(err, domain.ErrCouponStatus):
		c.JSON(http.StatusConflict, gin.H{"code": 17004, "message": "coupon status invalid"})
	case errors.Is(err, domain.ErrCouponThreshold):
		c.JSON(http.StatusBadRequest, gin.H{"code": 17005, "message": "threshold not met"})
	case errors.Is(err, domain.ErrCouponExpired):
		c.JSON(http.StatusBadRequest, gin.H{"code": 17006, "message": "coupon expired"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "coupon request failed"})
	}
	return true
}
