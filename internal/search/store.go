package search

import "strings"

type document struct {
	ProductID    string `db:"product_id" json:"product_id"`
	CategoryID   string `db:"category_id" json:"category_id"`
	Title        string `db:"title" json:"title"`
	Subtitle     string `db:"subtitle" json:"subtitle"`
	Status       int    `db:"status" json:"status"`
	MinPriceCent int64  `db:"min_price_cent" json:"min_price_cent"`
}

func (h *Handler) upsertDocument(doc document) error {
	var category any
	if doc.CategoryID != "" {
		category = doc.CategoryID
	}
	_, err := h.DB.Exec(`
		INSERT INTO product_documents
		(product_id, category_id, title, subtitle, status, min_price_cent)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  category_id = VALUES(category_id),
		  title = VALUES(title),
		  subtitle = VALUES(subtitle),
		  status = VALUES(status),
		  min_price_cent = VALUES(min_price_cent)`,
		doc.ProductID, category, doc.Title, doc.Subtitle, doc.Status, doc.MinPriceCent,
	)
	return err
}

func (h *Handler) searchDocuments(q, categoryID string, priceMin, priceMax *int64, sort string, limit, offset int) ([]document, error) {
	where := []string{"status = 1"}
	args := make([]any, 0, 8)
	if q != "" {
		like := "%" + escapeLike(q) + "%"
		where = append(where, "(title LIKE ? OR subtitle LIKE ?)")
		args = append(args, like, like)
	}
	if categoryID != "" {
		where = append(where, "category_id = ?")
		args = append(args, categoryID)
	}
	if priceMin != nil {
		where = append(where, "min_price_cent >= ?")
		args = append(args, *priceMin)
	}
	if priceMax != nil {
		where = append(where, "min_price_cent <= ?")
		args = append(args, *priceMax)
	}
	orderBy := "updated_at DESC"
	switch sort {
	case "", "new":
	case "price_asc":
		orderBy = "min_price_cent ASC, product_id ASC"
	case "price_desc":
		orderBy = "min_price_cent DESC, product_id ASC"
	default:
		return nil, errBadQuery
	}
	query := `
		SELECT product_id, COALESCE(category_id, '') AS category_id, title, subtitle, status, min_price_cent
		FROM product_documents
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows := []document{}
	err := h.DB.Select(&rows, query, args...)
	if rows == nil {
		rows = []document{}
	}
	return rows, err
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
