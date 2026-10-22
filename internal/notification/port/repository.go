package port

import (
	"github.com/your-org/go-minimall/internal/notification/domain"
)

type Repository interface {
	Record(eventID, eventType, userID string, payload []byte) (domain.GinLog, error)
	ListMine(userID string) ([]domain.GinLog, error)
	MarkRead(id, userID string) error
}
