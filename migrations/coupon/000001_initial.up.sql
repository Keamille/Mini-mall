CREATE TABLE IF NOT EXISTS coupon_templates (
  id             CHAR(26)     NOT NULL,
  name           VARCHAR(128) NOT NULL,
  type           TINYINT      NOT NULL COMMENT '1满减 2折扣',
  threshold_cent BIGINT       NOT NULL DEFAULT 0,
  discount_cent  BIGINT       NOT NULL DEFAULT 0 COMMENT '满减减额（分）',
  discount_rate  INT          NOT NULL DEFAULT 0 COMMENT '折扣率万分比，如 8500=85折',
  total_count    INT          NOT NULL DEFAULT 0,
  claimed_count  INT          NOT NULL DEFAULT 0,
  per_user_limit INT          NOT NULL DEFAULT 1,
  valid_from     DATETIME(3)  NOT NULL,
  valid_to       DATETIME(3)  NOT NULL,
  status         TINYINT      NOT NULL DEFAULT 1 COMMENT '1启用 2停用',
  created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_coupon_templates_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_coupons (
  id              CHAR(26)    NOT NULL,
  template_id     CHAR(26)    NOT NULL,
  user_id         CHAR(26)    NOT NULL,
  status          TINYINT     NOT NULL DEFAULT 1 COMMENT '1未用 2锁定 3已用 4过期',
  locked_order_id CHAR(26)    NULL,
  used_at         DATETIME(3) NULL,
  created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_user_coupons_user (user_id),
  KEY idx_user_coupons_template (template_id),
  CONSTRAINT fk_user_coupons_template
    FOREIGN KEY (template_id) REFERENCES coupon_templates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
