package app

import (
	"context"

	"github.com/your-org/go-minimall/internal/search/domain"
	"github.com/your-org/go-minimall/internal/search/port"
)

// Service is the search use-case layer.
type Service struct {
	Index port.Index
}

func NewService(index port.Index) *Service {
	return &Service{Index: index}
}

func (s *Service) Upsert(ctx context.Context, doc domain.Document) error {
	if len(doc.ProductID) != 26 || doc.Title == "" || doc.Status < 0 || doc.Status > 2 || doc.MinPriceCent < 0 {
		return domain.ErrBadDocument
	}
	return s.Index.Upsert(ctx, doc)
}

func (s *Service) Search(ctx context.Context, q domain.Query) ([]domain.Document, error) {
	switch q.Sort {
	case "", "new", "price_asc", "price_desc":
	default:
		return nil, domain.ErrBadQuery
	}
	return s.Index.Search(ctx, q)
}
