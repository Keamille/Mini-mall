package app

import (
	"database/sql"
	"errors"
	"github.com/your-org/go-minimall/internal/order/domain"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrOrderCannotPay = errors.New("order cannot be paid")

func (h *Handler) CreateOrder(c *gin.Context) {
	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "idempotency key is required"})
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 10002, "message": "user not found"})
		return
	}
	if existing, err := h.Repo.FindByIdempotency(userID, key); err == nil {
		payload := map[string]any{"pay_amount_cent": existing.PayAmountCent}
		h.Notify.Notify("order_created", existing.ID, existing.UserID, payload)
		h.Events.Publish("order_created", existing.ID, existing.UserID, payload)
		c.JSON(200, gin.H{"code": 0, "data": existing})
		return
	}
	var req struct {
		Items        []domain.ItemRow `json:"items" binding:"required"`
		UserCouponID string           `json:"user_coupon_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	var itemAmount int64
	for i := range req.Items {
		if req.Items[i].Quantity <= 0 || req.Items[i].SkuID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid item"})
			return
		}
		req.Items[i].AmountCent = req.Items[i].PriceCent * int64(req.Items[i].Quantity)
		itemAmount += req.Items[i].AmountCent
	}
	items := req.Items

	oderID := domain.NewID()
	var discount int64
	if req.UserCouponID != "" {
		d, status, code, message, err := h.Coupons.Lock(userID, req.UserCouponID, oderID, itemAmount)
		if err != nil || status != http.StatusOK {
			if status == 0 || err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to lock coupon"})
				return
			}
			c.JSON(status, gin.H{"code": code, "message": message})
			return
		}
		discount = d
	}
	payAmount := itemAmount - discount

	resevationID, statusCode, err := h.Inventory.Reserve(oderID, items)
	if statusCode == 409 {
		if req.UserCouponID != "" {
			h.Coupons.Unlock(userID, req.UserCouponID, oderID)
		}
		c.JSON(http.StatusConflict, gin.H{"code": 13001, "message": "insufficient stock"})
		return
	}
	if statusCode != 200 || err != nil {
		if req.UserCouponID != "" {
			h.Coupons.Unlock(userID, req.UserCouponID, oderID)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to reserve"})
		return
	}
	couponID := sql.NullString{}
	if req.UserCouponID != "" {
		couponID = sql.NullString{String: req.UserCouponID, Valid: true}
	}
	payExpire := newPayExpireAt()
	if err := h.Repo.InsertOrderTx(domain.OrderRow{
		ID:             oderID,
		OrderNo:        oderID,
		UserID:         userID,
		Status:         10,
		PayAmountCent:  payAmount,
		ItemAmountCent: itemAmount,
		DiscountCent:   discount,
		CouponID:       couponID,
		ReservationID:  sql.NullString{String: resevationID, Valid: true},
		IdempotencyKey: key,
		PayExpireAt:    payExpire,
	}, items); err != nil {
		h.Inventory.Release(oderID)
		if req.UserCouponID != "" {
			h.Coupons.Unlock(userID, req.UserCouponID, oderID)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create order"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"data": gin.H{
			"id":               oderID,
			"order_no":         oderID,
			"status":           10,
			"pay_amount_cent":  payAmount,
			"item_amount_cent": itemAmount,
			"discount_cent":    discount,
			"coupon_id":        req.UserCouponID,
			"reservation_id":   resevationID,
			"pay_expire_at":    payExpire.Time.UTC().Format(time.RFC3339Nano),
		},
	})
	payload := map[string]any{"pay_amount_cent": payAmount}
	h.Notify.Notify("order_created", oderID, userID, payload)
	h.Events.Publish("order_created", oderID, userID, payload)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	o, err := h.Repo.FindOrderByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 15001, "message": "order not found"})
		return
	}
	if o.Status != 10 {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be cancelled"})
		return
	}
	if err := h.Inventory.Release(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to release stock"})
		return
	}
	if o.CouponID.Valid && o.CouponID.String != "" {
		if err := h.Coupons.Unlock(userID, o.CouponID.String, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to unlock coupon"})
			return
		}
	}
	ok, err := h.Repo.MarkCancelled(id, userID)
	if err != nil || !ok {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be cancelled"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"id": id, "status": 30}})
	h.Notify.Notify("order_cancelled", id, userID, nil)
}

func (h *Handler) MarkPaid(c *gin.Context) {
	id := c.Param("id")
	if err := h.ApplyPaid(id); err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 15001, "message": "order not found"})
			return
		}
		if errors.Is(err, ErrOrderCannotPay) {
			c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be paid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "status": 20}})
}

func (h *Handler) ApplyPaid(id string) error {
	o, err := h.Repo.FindOrder(id)
	if err != nil {
		return ErrOrderNotFound
	}
	if o.Status == 20 {
		h.Notify.Notify("order_paid", id, o.UserID, nil)
		return nil
	}
	if o.Status != 10 {
		return ErrOrderCannotPay
	}
	if err := h.Inventory.Commit(id); err != nil {
		return errors.New("failed to commit stock")
	}
	if o.CouponID.Valid && o.CouponID.String != "" {
		if err := h.Coupons.Confirm(o.UserID, o.CouponID.String, id); err != nil {
			return errors.New("failed to confirm coupon")
		}
	}
	if _, err := h.Repo.MarkPaid(id); err != nil {
		return errors.New("failed to mark paid")
	}
	h.Notify.Notify("order_paid", id, o.UserID, nil)
	return nil
}
