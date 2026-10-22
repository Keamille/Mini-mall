package domain

import "errors"

const (
	// PhysicalIndex is the concrete OpenSearch index behind the read/write alias.
	PhysicalIndex = "minimall_products_v1"
	// Alias is the stable name callers search and upsert against.
	Alias = "minimall_products"
)

var (
	ErrBadQuery    = errors.New("bad search query")
	ErrBadDocument = errors.New("bad search document")
)

// Document is one product in the search index.
type Document struct {
	ProductID    string `json:"product_id"`
	CategoryID   string `json:"category_id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Status       int    `json:"status"`
	MinPriceCent int64  `json:"min_price_cent"`
}

// Query is a catalog search. Sort is "", "new", "price_asc", or "price_desc".
type Query struct {
	Text       string
	CategoryID string
	PriceMin   *int64
	PriceMax   *int64
	Sort       string
	Limit      int
	Offset     int
}
