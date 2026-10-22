package domain

import (
	"database/sql"
	"encoding/json"
)

type CategoryRow struct {
	ID       string         `db:"id" json:"id"`
	ParentID sql.NullString `db:"parent_id" json:"parent_id"`
	Name     string         `db:"name" json:"name"`
	Sort     int            `db:"sort" json:"sort"`
	Status   int8           `db:"status" json:"status"`
}

type SkuRow struct {
	ID        string          `db:"id" json:"id"`
	ProductID string          `db:"product_id" json:"product_id"`
	SkuCode   string          `db:"sku_code" json:"sku_code"`
	AttrsJSON json.RawMessage `db:"attrs_json" json:"attrs"`
	PriceCent int64           `db:"price_cent" json:"price_cent"`
	Status    int8            `db:"status" json:"status"`
}

type ProductRow struct {
	ID          string         `db:"id" json:"id"`
	CategoryID  sql.NullString `db:"category_id" json:"category_id"`
	Title       string         `db:"title" json:"title"`
	Subtitle    string         `db:"subtitle" json:"subtitle"`
	Description sql.NullString `db:"description" json:"description"`
	Status      int8           `db:"status" json:"status"`
}
