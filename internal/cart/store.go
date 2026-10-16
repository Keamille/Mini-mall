package cart

type cartItemRow struct {
	ID       string `db:"id" json:"id"`
	UserID   string `db:"user_id" json:"user_id"`
	SkuID    string `db:"sku_id" json:"sku_id"`
	Quantity int    `db:"quantity" json:"quantity"`
	Checked  int8   `db:"checked" json:"checked"`
}

func (h *Handler) listByuser(userID string) ([]cartItemRow, error) {
	var items []cartItemRow
	err := h.DB.Select(&items, `
		SELECT id, user_id, sku_id, quantity, checked
		FROM cart_items
		WHERE user_id = ?`, userID)
	return items, err
}

func (h *Handler) upsertItem(userID, skuID string, qty int) error {
	_, err := h.DB.Exec(`
		INSERT INTO cart_items (id, user_id, sku_id, quantity, checked)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE quantity = ?, checked = ?`,
		NewID(), userID, skuID, qty, 1, qty, 1)
	return err
}

func (h *Handler) deleteItem(userID, skuID string) error {
	_, err := h.DB.Exec(`
		DELETE FROM cart_items
		WHERE user_id = ? AND sku_id = ?`, userID, skuID)
	return err
}

func (h *Handler) updateChecked(userID string, skuID string, checked int8) error {
	_, err := h.DB.Exec(`
		UPDATE cart_items
		SET checked = ?
		WHERE user_id = ? AND sku_id = ?`, checked, userID, skuID)
	return err
}
