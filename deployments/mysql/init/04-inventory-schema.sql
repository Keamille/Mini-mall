-- inventory：库存与预留（对齐 docs/08-data-models.md）
USE minimall_inventory;

CREATE TABLE IF NOT EXISTS stocks (
  sku_id     CHAR(26)    NOT NULL,
  available  INT         NOT NULL DEFAULT 0,
  reserved   INT         NOT NULL DEFAULT 0,
  version    INT         NOT NULL DEFAULT 0,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (sku_id),
  CONSTRAINT chk_stocks_available CHECK (available >= 0),
  CONSTRAINT chk_stocks_reserved CHECK (reserved >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_reservations (
  id         CHAR(26)    NOT NULL,
  order_id   CHAR(26)    NOT NULL,
  status     TINYINT     NOT NULL DEFAULT 1 COMMENT '1预留 2已确认 3已释放',
  expires_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_reservations_order (order_id),
  KEY idx_stock_reservations_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_reservation_items (
  id             CHAR(26) NOT NULL,
  reservation_id CHAR(26) NOT NULL,
  sku_id         CHAR(26) NOT NULL,
  quantity       INT      NOT NULL,
  PRIMARY KEY (id),
  KEY idx_sri_reservation (reservation_id),
  KEY idx_sri_sku (sku_id),
  CONSTRAINT fk_sri_reservation
    FOREIGN KEY (reservation_id) REFERENCES stock_reservations (id) ON DELETE CASCADE,
  CONSTRAINT chk_sri_quantity CHECK (quantity > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_ledger (
  id          CHAR(26)     NOT NULL,
  sku_id      CHAR(26)     NOT NULL,
  delta       INT          NOT NULL COMMENT '增减数量（可负）',
  reason      VARCHAR(64)  NOT NULL DEFAULT '',
  ref_type    VARCHAR(32)  NOT NULL DEFAULT '' COMMENT 'order/reservation/...',
  ref_id      CHAR(26)     NULL,
  created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_stock_ledger_sku (sku_id),
  KEY idx_stock_ledger_ref (ref_type, ref_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
