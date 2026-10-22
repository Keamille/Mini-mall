package mysqlrepo

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/product/domain"
	"github.com/your-org/go-minimall/internal/product/port"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) ListCategories() ([]domain.CategoryRow, error) {
	var rows []domain.CategoryRow
	err := r.DB.Select(&rows, `
		SELECT id, parent_id, name, sort, status
		FROM categories WHERE status = 1
		ORDER BY sort ASC, name ASC`)
	return rows, err
}

func (r *Repo) CreateCategory(c domain.CategoryRow) error {
	var parent any
	if c.ParentID.Valid {
		parent = c.ParentID.String
	}
	_, err := r.DB.Exec(`
		INSERT INTO categories (id, parent_id, name, sort, status)
		VALUES (?, ?, ?, ?, ?)`,
		c.ID, parent, c.Name, c.Sort, c.Status,
	)
	return err
}

func (r *Repo) CreateProduct(p domain.ProductRow) error {
	var categoryID any
	if p.CategoryID.Valid {
		categoryID = p.CategoryID.String
	}
	var description any
	if p.Description.Valid {
		description = p.Description.String
	}
	_, err := r.DB.Exec(`
		INSERT INTO products (id, category_id, title, subtitle, description, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, categoryID, p.Title, p.Subtitle, description, p.Status,
	)
	return err
}

func (r *Repo) ListProducts(onlyOnSale bool, limit, offset int) ([]domain.ProductRow, error) {
	var rows []domain.ProductRow
	if onlyOnSale {
		err := r.DB.Select(&rows, `
			SELECT id, category_id, title, subtitle, description, status
			FROM products WHERE status = 1
			ORDER BY updated_at DESC
			LIMIT ? OFFSET ?`, limit, offset)
		return rows, err
	}
	err := r.DB.Select(&rows, `
		SELECT id, category_id, title, subtitle, description, status
		FROM products
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	return rows, err
}

func (r *Repo) FindProductByID(id string) (domain.ProductRow, error) {
	var p domain.ProductRow
	err := r.DB.Get(&p, `
		SELECT id, category_id, title, subtitle, description, status
		FROM products WHERE id = ?`, id)
	return p, err
}

func (r *Repo) ListSkusByProductID(productID string) ([]domain.SkuRow, error) {
	var rows []domain.SkuRow
	err := r.DB.Select(&rows, `
		SELECT id, product_id, sku_code, attrs_json, price_cent, status
		FROM skus WHERE product_id = ?
		ORDER BY created_at ASC`, productID)
	return rows, err
}

func (r *Repo) CreateSku(s domain.SkuRow) error {
	attrs := s.AttrsJSON
	if len(attrs) == 0 {
		attrs = json.RawMessage(`{}`)
	}
	_, err := r.DB.Exec(`
		INSERT INTO skus (id, product_id, sku_code, attrs_json, price_cent, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.ProductID, s.SkuCode, attrs, s.PriceCent, s.Status,
	)
	return err
}
