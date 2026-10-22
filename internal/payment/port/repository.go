package port

import (
	"github.com/your-org/go-minimall/internal/payment/domain"
)

type Repository interface {
	FindByOrderID(orderID string) (domain.PaymentRow, error)
	MarkPaid(id, user_id string) error
	InsertPayment(p domain.PaymentRow) error
	FindByID(id, userID string) (domain.PaymentRow, error)
}
