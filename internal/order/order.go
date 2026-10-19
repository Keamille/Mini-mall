package order

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errOrderNotFound = errors.New("order not found")
var errOrderCannotPay = errors.New("order cannot be paid")

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
	if existing, err := h.findByIdempotency(userID, key); err == nil {
		notifyOrder("order_created", existing.ID, existing.UserID, map[string]any{
			"pay_amount_cent": existing.PayAmountCent,
		})
		c.JSON(200, gin.H{"code": 0, "data": existing})
		return
	}
	var req struct {
		Items        []itemRow `json:"items" binding:"required"`
		UserCouponID string    `json:"user_coupon_id"`
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

	oderID := NewID()
	var discount int64
	if req.UserCouponID != "" {
		d, status, code, message, err := lockCoupon(userID, req.UserCouponID, oderID, itemAmount)
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

	resevationID, statusCode, err := reserve(oderID, items)
	if statusCode == 409 {
		if req.UserCouponID != "" {
			unlockCoupon(userID, req.UserCouponID, oderID)
		}
		c.JSON(http.StatusConflict, gin.H{"code": 13001, "message": "insufficient stock"})
		return
	}
	if statusCode != 200 || err != nil {
		if req.UserCouponID != "" {
			unlockCoupon(userID, req.UserCouponID, oderID)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to reserve"})
		return
	}
	couponID := sql.NullString{}
	if req.UserCouponID != "" {
		couponID = sql.NullString{String: req.UserCouponID, Valid: true}
	}
	if err := h.insertOrderTx(orderRow{
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
	}, items); err != nil {
		release(oderID)
		if req.UserCouponID != "" {
			unlockCoupon(userID, req.UserCouponID, oderID)
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
		},
	})
	notifyOrder("order_created", oderID, userID, map[string]any{
		"pay_amount_cent": payAmount,
	})
}

func (h *Handler) CancelOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	o, err := h.findOrderByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 15001, "message": "order not found"})
		return
	}
	if o.Status != 10 {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be cancelled"})
		return
	}
	if err := release(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to release stock"})
		return
	}
	if o.CouponID.Valid && o.CouponID.String != "" {
		if err := unlockCoupon(userID, o.CouponID.String, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to unlock coupon"})
			return
		}
	}
	ok, err := h.markCancelled(id, userID)
	if err != nil || !ok {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be cancelled"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"id": id, "status": 30}})
	notifyOrder("order_cancelled", id, userID, nil)
}

func (h *Handler) MarkPaid(c *gin.Context) {
	id := c.Param("id")
	if err := h.applyPaid(id); err != nil {
		if errors.Is(err, errOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 15001, "message": "order not found"})
			return
		}
		if errors.Is(err, errOrderCannotPay) {
			c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be paid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "status": 20}})
}

func (h *Handler) applyPaid(id string) error {
	o, err := h.findOrder(id)
	if err != nil {
		return errOrderNotFound
	}
	if o.Status == 20 {
		notifyOrder("order_paid", id, o.UserID, nil)
		return nil
	}
	if o.Status != 10 {
		return errOrderCannotPay
	}
	if err := commitStock(id); err != nil {
		return errors.New("failed to commit stock")
	}
	if o.CouponID.Valid && o.CouponID.String != "" {
		if err := confirmCoupon(o.UserID, o.CouponID.String, id); err != nil {
			return errors.New("failed to confirm coupon")
		}
	}
	if _, err := h.markPaid(id); err != nil {
		return errors.New("failed to mark paid")
	}
	notifyOrder("order_paid", id, o.UserID, nil)
	return nil
}
