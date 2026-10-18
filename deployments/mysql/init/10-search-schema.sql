-- search：商品检索文档。学习阶段用 MySQL 代替 OpenSearch。
USE minimall_search;

CREATE TABLE IF NOT EXISTS product_documents (
  product_id     CHAR(26)     NOT NULL,
  category_id    CHAR(26)     NULL,
  title          VARCHAR(255) NOT NULL,
  subtitle       VARCHAR(255) NOT NULL DEFAULT '',
  status         TINYINT      NOT NULL DEFAULT 0 COMMENT '0草稿 1上架 2下架',
  min_price_cent BIGINT       NOT NULL DEFAULT 0,
  updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (product_id),
  KEY idx_product_documents_status (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
