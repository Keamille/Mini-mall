package payment

type paymentRow struct {
	ID             string `db:"id" json:"id"`
	PaymentNo      string `db:"payment_no" json:"payment_no"`
	OrderID        string `db:"order_id" json:"order_id"`
	UserID         string `db:"user_id" json:"user_id"`
	AmountCent     int64  `db:"amount_cent" json:"amount_cent"`
	Status         int8   `db:"status" json:"status"`
	IdempotencyKey string `db:"idempotency_key" json:"idempotency_key"`
}

func (h *Handler) findByOrderID(orderID string) (paymentRow, error) {
	var p paymentRow
	err := h.DB.Get(&p, `
		SELECT id, payment_no, order_id, user_id, amount_cent, status, idempotency_key
		FROM payments WHERE order_id = ?`, orderID)
	return p, err
}

func (h *Handler) markPaid(id, user_id string) error {
	_, err := h.DB.Exec(`
		UPDATE payments
		SET status = 20, paid_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND user_id = ? AND status = 10`, id, user_id)
	return err
}

func (h *Handler) insertPayment(p paymentRow) error {
	_, err := h.DB.Exec(`
		INSERT INTO payments (id, payment_no, order_id, user_id, amount_cent, status, provider, idempotency_key)
		VALUES (?, ?, ?, ?, ?, 10, 'mock', ?)`,
		p.ID, p.PaymentNo, p.OrderID, p.UserID, p.AmountCent, p.IdempotencyKey,
	)
	return err
}

func (h *Handler) findByID(id, userID string) (paymentRow, error) {
	var p paymentRow
	err := h.DB.Get(&p, `
		SELECT id, payment_no, order_id, user_id, amount_cent, status, idempotency_key
		FROM payments WHERE id = ? AND user_id = ?`, id, userID)
	return p, err
}
