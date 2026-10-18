package product

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var searchClient = &http.Client{Timeout: 2 * time.Second}

func (h *Handler) indexProduct(p productRow) {
	var minPrice sql.NullInt64
	_ = h.DB.Get(&minPrice, `
		SELECT MIN(price_cent) FROM skus WHERE product_id = ? AND status = 1`, p.ID)
	body := map[string]any{
		"product_id":     p.ID,
		"title":          p.Title,
		"subtitle":       p.Subtitle,
		"status":         p.Status,
		"min_price_cent": minPrice.Int64,
	}
	if p.CategoryID.Valid {
		body["category_id"] = p.CategoryID.String
	}
	raw, _ := json.Marshal(body)
	resp, err := searchClient.Post("http://127.0.0.1:8084/api/v1/search/products", "application/json", bytes.NewReader(raw))
	if err != nil {
		log.Printf("search index %s: %v", p.ID, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("search index %s: status %d", p.ID, resp.StatusCode)
	}
}
