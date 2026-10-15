-- payment：支付单 / 退款（对齐 docs/08-data-models.md）
USE minimall_payment;

CREATE TABLE IF NOT EXISTS payments (
  id                CHAR(26)    NOT NULL,
  payment_no        VARCHAR(32) NOT NULL,
  order_id          CHAR(26)    NOT NULL,
  user_id           CHAR(26)    NOT NULL,
  amount_cent       BIGINT      NOT NULL,
  status            TINYINT     NOT NULL DEFAULT 10 COMMENT '10待支付 20成功 30失败 40关闭',
  provider          VARCHAR(32) NOT NULL DEFAULT 'mock',
  provider_trade_no VARCHAR(64) NULL,
  idempotency_key   VARCHAR(64) NOT NULL,
  paid_at           DATETIME(3) NULL,
  created_at        DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at        DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_payments_payment_no (payment_no),
  UNIQUE KEY uk_payments_order (order_id),
  UNIQUE KEY uk_payments_user_idempotency (user_id, idempotency_key),
  KEY idx_payments_user (user_id),
  KEY idx_payments_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS payment_refunds (
  id                 CHAR(26)    NOT NULL,
  payment_id         CHAR(26)    NOT NULL,
  amount_cent        BIGINT      NOT NULL,
  status             TINYINT     NOT NULL DEFAULT 10 COMMENT '10处理中 20成功 30失败',
  provider_refund_no VARCHAR(64) NULL,
  created_at         DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at         DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_payment_refunds_payment (payment_id),
  CONSTRAINT fk_payment_refunds_payment
    FOREIGN KEY (payment_id) REFERENCES payments (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
