package mysqlrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/cart/domain"
	"github.com/your-org/go-minimall/internal/cart/port"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) ListByuser(userID string) ([]domain.CartItemRow, error) {
	var items []domain.CartItemRow
	err := r.DB.Select(&items, `
		SELECT id, user_id, sku_id, quantity, checked
		FROM cart_items
		WHERE user_id = ?`, userID)
	return items, err
}

func (r *Repo) UpsertItem(userID, skuID string, qty int) error {
	_, err := r.DB.Exec(`
		INSERT INTO cart_items (id, user_id, sku_id, quantity, checked)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE quantity = ?, checked = ?`,
		domain.NewID(), userID, skuID, qty, 1, qty, 1)
	return err
}

func (r *Repo) DeleteItem(userID, skuID string) error {
	_, err := r.DB.Exec(`
		DELETE FROM cart_items
		WHERE user_id = ? AND sku_id = ?`, userID, skuID)
	return err
}

func (r *Repo) updateChecked(userID string, skuID string, checked int8) error {
	_, err := r.DB.Exec(`
		UPDATE cart_items
		SET checked = ?
		WHERE user_id = ? AND sku_id = ?`, checked, userID, skuID)
	return err
}
