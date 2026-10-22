package port

import (
	"github.com/your-org/go-minimall/internal/order/domain"
)

type Repository interface {
	FindByIdempotency(userID, key string) (domain.OrderRow, error)
	InsertOrderTx(o domain.OrderRow, items []domain.ItemRow) error
	FindOrderByID(id string, userID string) (domain.OrderRow, error)
	MarkCancelled(id string, userID string) (bool, error)
	FindOrder(id string) (domain.OrderRow, error)
	MarkPaid(id string) (bool, error)
	ListExpiredPending(limit int) ([]domain.OrderRow, error)
	MarkClosed(id string) (bool, error)
}
