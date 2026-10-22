package app

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	grpcadp "github.com/your-org/go-minimall/internal/payment/adapter/grpc"
	"github.com/your-org/go-minimall/internal/payment/domain"
)

type CreatePaymentRequest struct {
	OrderID    string `json:"order_id" binding:"required"`
	AmountCent int64  `json:"amount_cent" binding:"required"`
}

func (h *Handler) CreatePayment(c *gin.Context) {
	userID := c.GetString("user_id")
	key := c.GetHeader("Idempotency-Key")
	if userID == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and idempotency_key are required"})
		return
	}

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existing, err := h.Repo.FindByOrderID(req.OrderID); err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": existing})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to query payment"})
		return
	}

	p := domain.PaymentRow{
		ID:             domain.NewID(),
		OrderID:        req.OrderID,
		UserID:         userID,
		AmountCent:     req.AmountCent,
		Status:         10,
		IdempotencyKey: key,
	}
	p.PaymentNo = p.ID
	if err := h.Repo.InsertPayment(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": p})
}

func (h *Handler) MockConfirm(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	p, err := h.Repo.FindByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 16001, "message": "payment not found"})
		return
	}
	if p.Status != 20 {
		if err := h.Repo.MarkPaid(id, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to confirm"})
			return
		}
		p.Status = 20
	}
	if err := grpcadp.NotifyOrderPaid(p.OrderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to mark order paid"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}
