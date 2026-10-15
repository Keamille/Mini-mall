# MySQL 初始化脚本

仅在**数据卷为空**（首次 `up` 或 `down -v` 之后）由容器自动执行 `/docker-entrypoint-initdb.d`。

| 文件 | 库 | 表 |
|------|----|----|
| `01-create-databases.sql` | — | 创建各库并授权 |
| `02-user-schema.sql` | `minimall_user` | users, user_addresses, refresh_tokens |
| `03-product-schema.sql` | `minimall_product` | categories, products, skus |
| `04-inventory-schema.sql` | `minimall_inventory` | stocks, stock_reservations, stock_reservation_items, stock_ledger |
| `05-cart-schema.sql` | `minimall_cart` | cart_items |
| `06-order-schema.sql` | `minimall_order` | orders, order_items, outbox_events |
| `07-payment-schema.sql` | `minimall_payment` | payments, payment_refunds |
| `08-coupon-schema.sql` | `minimall_coupon` | coupon_templates, user_coupons |
| `09-notification-schema.sql` | `minimall_notification` | notification_logs |

字段定义对齐 [docs/08-data-models.md](../../../docs/08-data-models.md)；手工 migrate 见 `migrations/<svc>/`。

## 已有数据卷：补建新表（不清空数据）

```powershell
# Docker Desktop 已启动后：
Get-ChildItem deployments\mysql\init\0*.sql | Sort-Object Name | ForEach-Object {
  Get-Content $_.FullName -Raw | docker compose -f deployments/docker-compose.yml exec -T mysql mysql -uminimall -p123456
}
```

## 清空重来（做法 B）

```powershell
docker compose -f deployments/docker-compose.yml down -v
docker compose -f deployments/docker-compose.yml up -d mysql
```
