ALTER TABLE orders
  DROP KEY idx_orders_pay_expire,
  DROP COLUMN pay_expire_at;
