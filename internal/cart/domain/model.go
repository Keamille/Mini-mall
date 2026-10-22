package domain

type CartItemRow struct {
	ID       string `db:"id" json:"id"`
	UserID   string `db:"user_id" json:"user_id"`
	SkuID    string `db:"sku_id" json:"sku_id"`
	Quantity int    `db:"quantity" json:"quantity"`
	Checked  int8   `db:"checked" json:"checked"`
}
