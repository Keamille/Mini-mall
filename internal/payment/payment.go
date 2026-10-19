package payment

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	orderv1 "github.com/your-org/go-minimall/api/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	if existing, err := h.findByOrderID(req.OrderID); err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": existing})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to query payment"})
		return
	}

	p := paymentRow{
		ID:             NewID(),
		OrderID:        req.OrderID,
		UserID:         userID,
		AmountCent:     req.AmountCent,
		Status:         10,
		IdempotencyKey: key,
	}
	p.PaymentNo = p.ID
	if err := h.insertPayment(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": p})
}

func (h *Handler) MockConfirm(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	p, err := h.findByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 16001, "message": "payment not found"})
		return
	}
	if p.Status != 20 {
		if err := h.markPaid(id, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to confirm"})
			return
		}
		p.Status = 20
	}
	if err := notifyOrderPaid(p.OrderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to mark order paid"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}

func notifyOrderPaid(orderID string) error {
	conn, err := grpc.NewClient("127.0.0.1:9004", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = orderv1.NewOrderServiceClient(conn).MarkPaid(ctx, &orderv1.MarkPaidRequest{OrderId: orderID})
	if err != nil {
		return errors.New("order paid failed")
	}
	return nil
}
