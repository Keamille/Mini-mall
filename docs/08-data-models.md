# 08 数据模型

金额单位均为**分**（`BIGINT`）。时间字段 `DATETIME(3)` 或 `TIMESTAMP(3)`，UTC。主键/业务 ID 统一 **ULID**（`CHAR(26)`），见 [05-coding-standards.md](05-coding-standards.md)。下列为各服务核心表；完整约束以实现 migration 为准。

## 1. 库划分

| Database | 服务 |
|----------|------|
| `minimall_user` | user |
| `minimall_product` | product |
| `minimall_inventory` | inventory |
| `minimall_cart` | cart |
| `minimall_order` | order |
| `minimall_payment` | payment |
| `minimall_coupon` | coupon |
| `minimall_notification` | notification |

搜索主数据在 OpenSearch；可用 MySQL 存索引水位/outbox（可选）。

## 2. user

### users

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT/CHAR PK | 用户 ID |
| email | VARCHAR(255) UNIQUE | 可空若用手机 |
| phone | VARCHAR(32) UNIQUE | 可空若用邮箱 |
| password_hash | VARCHAR(255) | argon2id/bcrypt |
| nickname | VARCHAR(64) | |
| status | TINYINT | 1正常 2禁用 |
| created_at / updated_at | DATETIME(3) | |

### user_addresses

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| user_id | FK | |
| receiver_name | VARCHAR(64) | |
| phone | VARCHAR(32) | |
| province / city / district | VARCHAR(64) | |
| detail | VARCHAR(255) | |
| is_default | TINYINT | |
| created_at / updated_at | | |

### refresh_tokens

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| user_id | | |
| token_hash | CHAR(64) UNIQUE | 只存哈希 |
| expires_at | | |
| revoked_at | NULL | |
| created_at | | |

## 3. product

### categories

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| parent_id | NULL | 树形 |
| name | VARCHAR(128) | |
| sort | INT | |
| status | TINYINT | |

### products (SPU)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| category_id | | |
| title | VARCHAR(255) | |
| subtitle | VARCHAR(255) | |
| description | TEXT | |
| status | TINYINT | 0草稿 1上架 2下架 |
| created_at / updated_at | | |

### skus

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| product_id | | |
| sku_code | VARCHAR(64) UNIQUE | |
| attrs_json | JSON | 规格属性 |
| price_cent | BIGINT | 售价 |
| status | TINYINT | |
| created_at / updated_at | | |

> 可售库存不在 product 库。

## 4. inventory

### stocks

| 字段 | 类型 | 说明 |
|------|------|------|
| sku_id | PK | |
| available | INT | 可预留数量 |
| reserved | INT | 已预留未确认 |
| version | INT | 乐观锁 |
| updated_at | | |

约束：`available >= 0`, `reserved >= 0`。

### stock_reservations

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | reservation_id |
| order_id | UNIQUE | 一单一预留单（可扩展多仓） |
| status | TINYINT | 1预留 2已确认 3已释放 |
| expires_at | | 超时释放 |
| created_at / updated_at | | |

### stock_reservation_items

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| reservation_id | | |
| sku_id | | |
| quantity | INT | |

### stock_ledger（可选）

流水审计：增减、原因、关联单号。

## 5. cart

### cart_items

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| user_id | | |
| sku_id | | |
| quantity | INT | |
| checked | TINYINT | 是否选中结算 |
| updated_at | | |

UNIQUE(`user_id`,`sku_id`)。

## 6. order

### orders

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | order_id |
| order_no | VARCHAR(32) UNIQUE | 展示号 |
| user_id | | |
| status | TINYINT | 见状态机 |
| pay_amount_cent | BIGINT | 应付 |
| item_amount_cent | BIGINT | 商品总额 |
| discount_cent | BIGINT | 优惠 |
| freight_cent | BIGINT | 运费一期可为 0 |
| currency | CHAR(3) | CNY |
| address_snapshot | JSON | 下单地址快照 |
| coupon_id | NULL | |
| reservation_id | NULL | |
| idempotency_key | VARCHAR(64) | UNIQUE(user_id, key) |
| paid_at / cancelled_at / created_at / updated_at | | |

订单状态建议：

| 值 | 名 | 说明 |
|----|----|------|
| 10 | PENDING_PAYMENT | 待支付 |
| 20 | PAID | 已支付 |
| 30 | CANCELLED | 已取消 |
| 40 | CLOSED | 超时关单 |

（发货/完成可二期扩展 50/60）

### order_items

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| order_id | | |
| sku_id | | |
| product_id | | |
| title | VARCHAR(255) | 快照 |
| sku_attrs_json | JSON | 快照 |
| price_cent | BIGINT | 快照单价 |
| quantity | INT | |
| amount_cent | BIGINT | |

### outbox_events（推荐）

本地消息表，保证与订单事务同库提交后异步发 NATS。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| aggregate_type / aggregate_id | | |
| event_type | | |
| payload | JSON | |
| status | | PENDING/SENT |
| created_at / sent_at | | |

## 7. payment

### payments

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| payment_no | UNIQUE | |
| order_id | UNIQUE | 一期一单一支付 |
| user_id | | |
| amount_cent | BIGINT | |
| status | TINYINT | 10待支付 20成功 30失败 40关闭 |
| provider | VARCHAR(32) | mock/wechat/alipay... |
| provider_trade_no | VARCHAR(64) | |
| idempotency_key | | |
| paid_at / created_at / updated_at | | |

### payment_refunds（二期可用）

退款单：amount、status、provider_refund_no。

## 8. coupon

### coupon_templates

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| name | | |
| type | TINYINT | 1满减 2折扣 |
| threshold_cent | BIGINT | 满减门槛 |
| discount_cent | BIGINT | 减额（满减） |
| discount_rate | INT | 折扣率万分比（可选） |
| total_count | INT | 发行量 |
| claimed_count | INT | |
| per_user_limit | INT | |
| valid_from / valid_to | | |
| status | | |

### user_coupons

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| template_id | | |
| user_id | | |
| status | TINYINT | 1未用 2锁定 3已用 4过期 |
| locked_order_id | NULL | |
| used_at | | |
| created_at | | |

## 9. notification

### notification_logs

| 字段 | 类型 | 说明 |
|------|------|------|
| id | PK | |
| user_id | | |
| channel | VARCHAR(16) | sms/email/inbox |
| template | | |
| payload | JSON | |
| status | | |
| event_id | UNIQUE | 幂等 |
| created_at | | |

## 10. 相关文档

- 事件：[09-async-events.md](09-async-events.md)
- 各服务：[06-services/](06-services/)
- 端口/库名：[14-port-matrix.md](14-port-matrix.md)

> 各服务规格中的「数据模型」仅作要点回链；**字段级定义以本文为准**。

---

导航：[上一篇 07 API 约定](07-api-conventions.md) | [下一篇 09 异步事件](09-async-events.md) | [返回索引](README.md)
