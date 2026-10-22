package port

import (
	"context"

	"github.com/your-org/go-minimall/internal/search/domain"
)

// Index is the search engine port. OpenSearch is the adapter.
type Index interface {
	Ping(ctx context.Context) error
	Ensure(ctx context.Context) error
	Upsert(ctx context.Context, doc domain.Document) error
	Search(ctx context.Context, q domain.Query) ([]domain.Document, error)
}
