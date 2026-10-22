-- 待支付订单支付截止时间，供超时关单扫描。
ALTER TABLE orders
  ADD COLUMN pay_expire_at DATETIME(3) NULL AFTER cancelled_at,
  ADD KEY idx_orders_pay_expire (status, pay_expire_at);
