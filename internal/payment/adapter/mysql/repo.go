package mysqlrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/payment/domain"
	"github.com/your-org/go-minimall/internal/payment/port"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) FindByOrderID(orderID string) (domain.PaymentRow, error) {
	var p domain.PaymentRow
	err := r.DB.Get(&p, `
		SELECT id, payment_no, order_id, user_id, amount_cent, status, idempotency_key
		FROM payments WHERE order_id = ?`, orderID)
	return p, err
}

func (r *Repo) MarkPaid(id, user_id string) error {
	_, err := r.DB.Exec(`
		UPDATE payments
		SET status = 20, paid_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND user_id = ? AND status = 10`, id, user_id)
	return err
}

func (r *Repo) InsertPayment(p domain.PaymentRow) error {
	_, err := r.DB.Exec(`
		INSERT INTO payments (id, payment_no, order_id, user_id, amount_cent, status, provider, idempotency_key)
		VALUES (?, ?, ?, ?, ?, 10, 'mock', ?)`,
		p.ID, p.PaymentNo, p.OrderID, p.UserID, p.AmountCent, p.IdempotencyKey,
	)
	return err
}

func (r *Repo) FindByID(id, userID string) (domain.PaymentRow, error) {
	var p domain.PaymentRow
	err := r.DB.Get(&p, `
		SELECT id, payment_no, order_id, user_id, amount_cent, status, idempotency_key
		FROM payments WHERE id = ? AND user_id = ?`, id, userID)
	return p, err
}
