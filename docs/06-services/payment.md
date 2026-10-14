# Payment Service

## 1. 职责与非职责

**职责**：创建支付单、渠道适配器（含 Mock）、回调验签、支付状态；退款抽象二期。

**非职责**：改订单明细/库存/券（通过事件通知）。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `payment-svc` |
| gRPC | `9005` |
| HTTP 回调 | `8085` |
| DB | `minimall_payment` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

- `POST /api/v1/payments/{id}/mock-confirm` — 仅 `MINIMALL_PAYMENT_MOCK=true` / `ENV=dev`
- `POST /api/v1/payments/callbacks/{provider}` — 验签后更新状态并发事件

## 4. 对内 gRPC

```protobuf
service PaymentService {
  rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
  rpc GetPayment(GetPaymentRequest) returns (Payment);
  rpc MockConfirm(MockConfirmRequest) returns (google.protobuf.Empty);
}
```

同一 `order_id` 一笔有效待支付单；幂等返回。

## 5. 数据模型

表：`payments`（可选 `payment_refunds`）。见 [08-data-models.md](../08-data-models.md#7-payment)。

## 6. 关键流程

CreatePayment → 返回 pay_params → 回调/Mock 成功 → outbox `PaymentSucceeded` → order 消费。

`PaymentProvider` 接口；一期 `MockProvider`。

## 7. 错误码

本域 `16001`–`16005`，见 [15-error-codes.md](../15-error-codes.md#8-payment16000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_PAYMENT_MYSQL_DSN` | |
| `MINIMALL_PAYMENT_GRPC_ADDR` | `:9005` |
| `MINIMALL_PAYMENT_HTTP_ADDR` | `:8085` |
| `MINIMALL_PAYMENT_MOCK` | `true` |
| `MINIMALL_PAYMENT_NATS_URL` | |

## 9. 依赖与事件

- 被 order 调 CreatePayment
- 生产：`PaymentSucceeded` 等

## 10. 相关文档

- [order.md](order.md)
- [12-security.md](../12-security.md)

---

导航：[上一篇 order](order.md) | [下一篇 coupon](coupon.md) | [服务索引](README.md)
