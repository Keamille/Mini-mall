package app

import (
	"database/sql"
	"github.com/your-org/go-minimall/internal/order/domain"
	"log"
	"os"
	"strconv"
	"time"
)

// PaymentTTL is how long a pending order may wait before auto-close (status 40).
func PaymentTTL() time.Duration {
	sec, _ := strconv.Atoi(os.Getenv("MINIMALL_ORDER_PAYMENT_TTL"))
	if sec <= 0 {
		sec = 1800
	}
	return time.Duration(sec) * time.Second
}

func closeInterval() time.Duration {
	sec, _ := strconv.Atoi(os.Getenv("MINIMALL_ORDER_CLOSE_INTERVAL_SEC"))
	if sec <= 0 {
		sec = 5
	}
	return time.Duration(sec) * time.Second
}

// StartCloseWorker periodically closes unpaid orders past pay_expire_at.
func (h *Handler) StartCloseWorker() {
	interval := closeInterval()
	log.Printf("order close worker every %s (payment TTL %s)", interval, PaymentTTL())
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			n, err := h.closeExpiredOrders()
			if err != nil {
				log.Printf("order close scan: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("order close worker closed %d order(s)", n)
			}
		}
	}()
}

func (h *Handler) closeExpiredOrders() (int, error) {
	list, err := h.Repo.ListExpiredPending(50)
	if err != nil {
		return 0, err
	}
	closed := 0
	for _, o := range list {
		if err := h.closeOne(o); err != nil {
			log.Printf("order close %s: %v", o.ID, err)
			continue
		}
		closed++
	}
	return closed, nil
}

// closeOne mirrors cancel: release stock, unlock coupon, status 40, notify.
func (h *Handler) closeOne(o domain.OrderRow) error {
	if err := h.Inventory.Release(o.ID); err != nil {
		return err
	}
	if o.CouponID.Valid && o.CouponID.String != "" {
		if err := h.Coupons.Unlock(o.UserID, o.CouponID.String, o.ID); err != nil {
			return err
		}
	}
	ok, err := h.Repo.MarkClosed(o.ID)
	if err != nil {
		return err
	}
	if !ok {
		// already paid/cancelled/closed by another path
		return nil
	}
	h.Notify.Notify("order_closed", o.ID, o.UserID, nil)
	return nil
}

func newPayExpireAt() sql.NullTime {
	return sql.NullTime{Time: time.Now().UTC().Add(PaymentTTL()), Valid: true}
}
