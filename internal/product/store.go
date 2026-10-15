package product

import (
	"database/sql"
	"encoding/json"
)

type categoryRow struct {
	ID       string         `db:"id" json:"id"`
	ParentID sql.NullString `db:"parent_id" json:"parent_id"`
	Name     string         `db:"name" json:"name"`
	Sort     int            `db:"sort" json:"sort"`
	Status   int8           `db:"status" json:"status"`
}
type skuRow struct {
	ID        string          `db:"id" json:"id"`
	ProductID string          `db:"product_id" json:"product_id"`
	SkuCode   string          `db:"sku_code" json:"sku_code"`
	AttrsJSON json.RawMessage `db:"attrs_json" json:"attrs"`
	PriceCent int64           `db:"price_cent" json:"price_cent"`
	Status    int8            `db:"status" json:"status"`
}
type productRow struct {
	ID          string         `db:"id" json:"id"`
	CategoryID  sql.NullString `db:"category_id" json:"category_id"`
	Title       string         `db:"title" json:"title"`
	Subtitle    string         `db:"subtitle" json:"subtitle"`
	Description sql.NullString `db:"description" json:"description"`
	Status      int8           `db:"status" json:"status"`
}

func (h *Handler) listCategories() ([]categoryRow, error) {
	var rows []categoryRow
	err := h.DB.Select(&rows, `
		SELECT id, parent_id, name, sort, status
		FROM categories WHERE status = 1
		ORDER BY sort ASC, name ASC`)
	return rows, err
}

func (h *Handler) createCategory(c categoryRow) error {
	var parent any
	if c.ParentID.Valid {
		parent = c.ParentID.String
	}
	_, err := h.DB.Exec(`
		INSERT INTO categories (id, parent_id, name, sort, status)
		VALUES (?, ?, ?, ?, ?)`,
		c.ID, parent, c.Name, c.Sort, c.Status,
	)
	return err
}

func (h *Handler) createProduct(p productRow) error {
	var categoryID any
	if p.CategoryID.Valid {
		categoryID = p.CategoryID.String
	}
	var description any
	if p.Description.Valid {
		description = p.Description.String
	}
	_, err := h.DB.Exec(`
		INSERT INTO products (id, category_id, title, subtitle, description, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, categoryID, p.Title, p.Subtitle, description, p.Status,
	)
	return err
}

func (h *Handler) listProducts(onlyOnSale bool, limit, offset int) ([]productRow, error) {
	var rows []productRow
	if onlyOnSale {
		err := h.DB.Select(&rows, `
			SELECT id, category_id, title, subtitle, description, status
			FROM products WHERE status = 1
			ORDER BY updated_at DESC
			LIMIT ? OFFSET ?`, limit, offset)
		return rows, err
	}
	err := h.DB.Select(&rows, `
		SELECT id, category_id, title, subtitle, description, status
		FROM products
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	return rows, err
}

func (h *Handler) findProductByID(id string) (productRow, error) {
	var p productRow
	err := h.DB.Get(&p, `
		SELECT id, category_id, title, subtitle, description, status
		FROM products WHERE id = ?`, id)
	return p, err
}

func (h *Handler) listSkusByProductID(productID string) ([]skuRow, error) {
	var rows []skuRow
	err := h.DB.Select(&rows, `
		SELECT id, product_id, sku_code, attrs_json, price_cent, status
		FROM skus WHERE product_id = ?
		ORDER BY created_at ASC`, productID)
	return rows, err
}

func (h *Handler) createSku(s skuRow) error {
	attrs := s.AttrsJSON
	if len(attrs) == 0 {
		attrs = json.RawMessage(`{}`)
	}
	_, err := h.DB.Exec(`
		INSERT INTO skus (id, product_id, sku_code, attrs_json, price_cent, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.ProductID, s.SkuCode, attrs, s.PriceCent, s.Status,
	)
	return err
}
