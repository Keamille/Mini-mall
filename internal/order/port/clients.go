package port

import "github.com/your-org/go-minimall/internal/order/domain"

type Inventory interface {
	Reserve(orderID string, items []domain.ItemRow) (string, int, error)
	Release(orderID string) error
	Commit(orderID string) error
}

type Coupons interface {
	Lock(userID, userCouponID, orderID string, itemAmount int64) (int64, int, int, string, error)
	Unlock(userID, userCouponID, orderID string) error
	Confirm(userID, userCouponID, orderID string) error
}

type Notifier interface {
	Notify(eventType, orderID, userID string, payload map[string]any)
}

type Publisher interface {
	Publish(eventType, orderID, userID string, payload map[string]any)
}
