-- user 服务初始表结构（对齐 docs/08-data-models.md）
-- 主键统一 ULID：CHAR(26)

CREATE TABLE IF NOT EXISTS users (
  id            CHAR(26)     NOT NULL,
  email         VARCHAR(255) NULL,
  phone         VARCHAR(32)  NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname      VARCHAR(64)  NOT NULL DEFAULT '',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '1正常 2禁用',
  created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_email (email),
  UNIQUE KEY uk_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_addresses (
  id            CHAR(26)     NOT NULL,
  user_id       CHAR(26)     NOT NULL,
  receiver_name VARCHAR(64)  NOT NULL,
  phone         VARCHAR(32)  NOT NULL,
  province      VARCHAR(64)  NOT NULL DEFAULT '',
  city          VARCHAR(64)  NOT NULL DEFAULT '',
  district      VARCHAR(64)  NOT NULL DEFAULT '',
  detail        VARCHAR(255) NOT NULL,
  is_default    TINYINT      NOT NULL DEFAULT 0,
  created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_user_addresses_user_id (user_id),
  CONSTRAINT fk_user_addresses_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id         CHAR(26)    NOT NULL,
  user_id    CHAR(26)    NOT NULL,
  token_hash CHAR(64)    NOT NULL,
  expires_at DATETIME(3) NOT NULL,
  revoked_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_refresh_tokens_token_hash (token_hash),
  KEY idx_refresh_tokens_user_id (user_id),
  CONSTRAINT fk_refresh_tokens_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
