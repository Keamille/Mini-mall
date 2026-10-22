package port

import (
	"github.com/your-org/go-minimall/internal/cart/domain"
)

type Repository interface {
	ListByuser(userID string) ([]domain.CartItemRow, error)
	UpsertItem(userID, skuID string, qty int) error
	DeleteItem(userID, skuID string) error
}
