package port

import (
	"time"

	"github.com/your-org/go-minimall/internal/inventory/domain"
)

type Repository interface {
	FindStock(skuID string) (domain.StockRow, error)
	InitStock(skuID string, available int) error
	ReserveTx(orderID string, expiresAt time.Time, items []domain.ReserveItem) (string, error)
	ReleaseTx(orderID string) error
	CommitTx(orderID string) error
}
