CREATE TABLE IF NOT EXISTS cart_items (
  id         CHAR(26)    NOT NULL,
  user_id    CHAR(26)    NOT NULL,
  sku_id     CHAR(26)    NOT NULL,
  quantity   INT         NOT NULL DEFAULT 1,
  checked    TINYINT     NOT NULL DEFAULT 1 COMMENT '是否选中结算',
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_cart_items_user_sku (user_id, sku_id),
  KEY idx_cart_items_user (user_id),
  CONSTRAINT chk_cart_items_qty CHECK (quantity > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
