# Coupon Service

## 1. 职责与非职责

**职责**：券模板、领券、下单锁定、支付核销、取消解锁、试算。

**非职责**：创建订单、运费、复杂促销叠加（一期单券）。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `coupon-svc` |
| gRPC | `9006` |
| DB | `minimall_coupon` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

- `GET /api/v1/coupons/templates`
- `POST /api/v1/coupons/templates/{id}/claim`（建议 Idempotency-Key）
- `GET /api/v1/coupons/mine`
- `POST /api/v1/coupons/quote` — 试算 `discount_cent`

## 4. 对内 gRPC

```protobuf
service CouponService {
  rpc Quote(QuoteRequest) returns (QuoteResponse);
  rpc LockCoupon(LockCouponRequest) returns (LockCouponResponse);
  rpc ConfirmUse(ConfirmUseRequest) returns (google.protobuf.Empty);
  rpc UnlockCoupon(UnlockCouponRequest) returns (google.protobuf.Empty);
}
```

一期满减：门槛 `threshold_cent`，减免 `discount_cent`（不超过商品额）。

## 5. 数据模型

表：`coupon_templates`、`user_coupons`。见 [08-data-models.md](../08-data-models.md#8-coupon)。

## 6. 关键流程

领券 → 下单 Lock → 支付成功 ConfirmUse / 取消 Unlock。

## 7. 错误码

本域 `17001`–`17006`，见 [15-error-codes.md](../15-error-codes.md#9-coupon17000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_COUPON_MYSQL_DSN` | |
| `MINIMALL_COUPON_GRPC_ADDR` | `:9006` |
| `MINIMALL_COUPON_NATS_URL` | |

## 9. 依赖与事件

- 被 order 同步调用
- 可选：`CouponLocked` / `CouponUsed` / `CouponUnlocked`

## 10. 相关文档

- [order.md](order.md)
- [08-data-models.md](../08-data-models.md)

---

导航：[上一篇 payment](payment.md) | [下一篇 search](search.md) | [服务索引](README.md)
