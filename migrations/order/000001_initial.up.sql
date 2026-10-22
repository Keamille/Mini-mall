CREATE TABLE IF NOT EXISTS orders (
  id               CHAR(26)     NOT NULL,
  order_no         VARCHAR(32)  NOT NULL,
  user_id          CHAR(26)     NOT NULL,
  status           TINYINT      NOT NULL DEFAULT 10 COMMENT '10待支付 20已支付 30已取消 40关闭',
  pay_amount_cent  BIGINT       NOT NULL DEFAULT 0,
  item_amount_cent BIGINT       NOT NULL DEFAULT 0,
  discount_cent    BIGINT       NOT NULL DEFAULT 0,
  freight_cent     BIGINT       NOT NULL DEFAULT 0,
  currency         CHAR(3)      NOT NULL DEFAULT 'CNY',
  address_snapshot JSON         NULL,
  coupon_id        CHAR(26)     NULL,
  reservation_id   CHAR(26)     NULL,
  idempotency_key  VARCHAR(64)  NOT NULL,
  paid_at          DATETIME(3)  NULL,
  cancelled_at     DATETIME(3)  NULL,
  pay_expire_at    DATETIME(3)  NULL COMMENT '待支付截止，超时关单 status=40',
  created_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_no (order_no),
  UNIQUE KEY uk_orders_user_idempotency (user_id, idempotency_key),
  KEY idx_orders_user (user_id),
  KEY idx_orders_status (status),
  KEY idx_orders_pay_expire (status, pay_expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_items (
  id             CHAR(26)     NOT NULL,
  order_id       CHAR(26)     NOT NULL,
  sku_id         CHAR(26)     NOT NULL,
  product_id     CHAR(26)     NOT NULL,
  title          VARCHAR(255) NOT NULL,
  sku_attrs_json JSON         NULL,
  price_cent     BIGINT       NOT NULL,
  quantity       INT          NOT NULL,
  amount_cent    BIGINT       NOT NULL,
  PRIMARY KEY (id),
  KEY idx_order_items_order (order_id),
  CONSTRAINT fk_order_items_order
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
  CONSTRAINT chk_order_items_qty CHECK (quantity > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS outbox_events (
  id             CHAR(26)     NOT NULL,
  aggregate_type VARCHAR(64)  NOT NULL,
  aggregate_id   CHAR(26)     NOT NULL,
  event_type     VARCHAR(64)  NOT NULL,
  payload        JSON         NOT NULL,
  status         TINYINT      NOT NULL DEFAULT 0 COMMENT '0PENDING 1SENT',
  created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  sent_at        DATETIME(3)  NULL,
  PRIMARY KEY (id),
  KEY idx_outbox_status_created (status, created_at),
  KEY idx_outbox_aggregate (aggregate_type, aggregate_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
