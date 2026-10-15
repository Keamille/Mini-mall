-- notification：通知日志（对齐 docs/08-data-models.md）
USE minimall_notification;

CREATE TABLE IF NOT EXISTS notification_logs (
  id         CHAR(26)     NOT NULL,
  user_id    CHAR(26)     NOT NULL,
  channel    VARCHAR(16)  NOT NULL COMMENT 'sms/email/inbox',
  template   VARCHAR(64)  NOT NULL DEFAULT '',
  payload    JSON         NULL,
  status     TINYINT      NOT NULL DEFAULT 0 COMMENT '0待发 1成功 2失败',
  event_id   VARCHAR(64)  NOT NULL COMMENT '幂等键',
  created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_notification_logs_event (event_id),
  KEY idx_notification_logs_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
