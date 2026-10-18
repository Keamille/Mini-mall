package order

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(200, gin.H{"code": 0, "data": existing})
		return
	}
	var req struct {
		Items []itemRow `json:"items" binding:"required"`
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
	resevationID, statusCode, err := reserve(oderID, items)
	if statusCode == 409 {
		c.JSON(http.StatusConflict, gin.H{"code": 13001, "message": "insufficient stock"})
		return
	}
	if statusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to reserve"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to reserve"})
		return
	}
	if err := h.insertOrderTx(orderRow{
		ID:             oderID,
		OrderNo:        oderID,
		UserID:         userID,
		Status:         10,
		PayAmountCent:  itemAmount,
		ItemAmountCent: itemAmount,
		ReservationID:  sql.NullString{String: resevationID, Valid: true},
		IdempotencyKey: key,
	}, items); err != nil {
		release(oderID)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create order"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"data": gin.H{
			"id":              oderID,
			"order_no":        oderID,
			"status":          10,
			"pay_amount_cent": itemAmount,
			"reservation_id":  resevationID,
		},
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
	ok, err := h.markCancelled(id, userID)
	if err != nil || !ok {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be cancelled"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"id": id, "status": 30}})
}

func (h *Handler) MarkPaid(c *gin.Context) {
	id := c.Param("id")
	o, err := h.findOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 15001, "message": "order not found"})
		return
	}
	if o.Status == 20 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "status": 20}})
		return
	}
	if o.Status != 10 {
		c.JSON(http.StatusConflict, gin.H{"code": 15002, "message": "order cannot be paid"})
		return
	}
	if err := commitStock(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to commit stock"})
		return
	}
	if _, err := h.markPaid(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to mark paid"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "status": 20}})
}
