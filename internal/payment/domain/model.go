package domain

type PaymentRow struct {
	ID             string `db:"id" json:"id"`
	PaymentNo      string `db:"payment_no" json:"payment_no"`
	OrderID        string `db:"order_id" json:"order_id"`
	UserID         string `db:"user_id" json:"user_id"`
	AmountCent     int64  `db:"amount_cent" json:"amount_cent"`
	Status         int8   `db:"status" json:"status"`
	IdempotencyKey string `db:"idempotency_key" json:"idempotency_key"`
}
