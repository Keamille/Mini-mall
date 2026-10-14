# Order Service

## 1. 职责与非职责

**职责**：下单编排、订单状态机、取消/超时关单、支付成功推进、订单查询。

**非职责**：直连第三方支付（走 payment）、库存真源（走 inventory）、券模板（走 coupon）。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `order-svc` |
| gRPC | `9004` |
| DB | `minimall_order` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

### POST `/api/v1/orders`

Headers：`Idempotency-Key` 必填。  
Body：`address_id`、`items[]` 或 `from_cart=true`、`user_coupon_id` 可选。  
响应：`order_id`、`order_no`、`status=10`、`pay_amount_cent`、`payment.pay_params`。

### GET `/api/v1/orders` · `GET /api/v1/orders/{id}`

### POST `/api/v1/orders/{id}/cancel`

仅 `PENDING_PAYMENT`；释放库存、解锁券。

## 4. 对内 gRPC / 事件消费

```protobuf
service OrderService {
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc MarkOrderPaid(MarkOrderPaidRequest) returns (google.protobuf.Empty);
}
```

优先经 NATS 消费 `PaymentSucceeded` 再内部推进（勿对公网暴露 Mark）。

## 5. 数据模型

表：`orders`、`order_items`、`outbox_events`。状态：10 待支付 / 20 已支付 / 30 取消 / 40 关单。见 [08-data-models.md](../08-data-models.md#6-order)。

## 6. 关键流程

**下单（同步）**：GetAddress → BatchGetSkus → LockCoupon? → ReserveStock → 写订单+outbox → CreatePayment → ClearCart?

**支付成功（异步）**：PaymentSucceeded → PAID + OrderPaid → CommitStock → ConfirmUse coupon

失败补偿：Lock 后失败 Unlock；Reserve 后写库失败 Release。

## 7. 错误码

本域 `15001`–`15006`，见 [15-error-codes.md](../15-error-codes.md#7-order15000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_ORDER_MYSQL_DSN` | |
| `MINIMALL_ORDER_GRPC_ADDR` | `:9004` |
| `MINIMALL_ORDER_PAYMENT_TTL` | 默认 1800 |
| `MINIMALL_ORDER_*_GRPC` | user/product/inventory/payment/coupon/cart |
| `MINIMALL_ORDER_NATS_URL` | |

## 9. 依赖与事件

- 同步：user, product, inventory, payment, coupon, cart
- 消费：`PaymentSucceeded`；生产：`OrderCreated` / `OrderPaid` / `OrderCancelled` / `OrderClosed`

## 10. 相关文档

- [inventory.md](inventory.md) · [payment.md](payment.md)
- [09-async-events.md](../09-async-events.md)

---

导航：[上一篇 cart](cart.md) | [下一篇 payment](payment.md) | [服务索引](README.md)
