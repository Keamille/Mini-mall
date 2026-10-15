-- product：类目 / SPU / SKU（对齐 docs/08-data-models.md）
USE minimall_product;

CREATE TABLE IF NOT EXISTS categories (
  id         CHAR(26)     NOT NULL,
  parent_id  CHAR(26)     NULL,
  name       VARCHAR(128) NOT NULL,
  sort       INT          NOT NULL DEFAULT 0,
  status     TINYINT      NOT NULL DEFAULT 1,
  PRIMARY KEY (id),
  KEY idx_categories_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
  id          CHAR(26)     NOT NULL,
  category_id CHAR(26)     NULL,
  title       VARCHAR(255) NOT NULL,
  subtitle    VARCHAR(255) NOT NULL DEFAULT '',
  description TEXT         NULL,
  status      TINYINT      NOT NULL DEFAULT 0 COMMENT '0草稿 1上架 2下架',
  created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_products_category (category_id),
  KEY idx_products_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS skus (
  id          CHAR(26)     NOT NULL,
  product_id  CHAR(26)     NOT NULL,
  sku_code    VARCHAR(64)  NOT NULL,
  attrs_json  JSON         NULL,
  price_cent  BIGINT       NOT NULL,
  status      TINYINT      NOT NULL DEFAULT 1,
  created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_skus_code (sku_code),
  KEY idx_skus_product (product_id),
  CONSTRAINT fk_skus_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
