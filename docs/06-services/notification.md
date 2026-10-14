# Notification Service

## 1. 职责与非职责

**职责**：订阅领域事件，按模板发送通知；一期以 **Log/Mock** 为主并落发送记录。

**非职责**：业务状态机、交易补偿。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `notification-svc` |
| gRPC | `9009`（运维补发） |
| DB | `minimall_notification` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST（可选）

- `GET /api/v1/notifications/mine`
- `POST /api/v1/notifications/{id}/read`

## 4. 对内 gRPC

```protobuf
service NotificationService {
  rpc Send(SendRequest) returns (google.protobuf.Empty);
}
```

普通路径走事件消费。

## 5. 数据模型

表：`notification_logs`（`event_id` 唯一幂等）。见 [08-data-models.md](../08-data-models.md#9-notification)。

## 6. 关键流程

| 事件 | 模板建议 |
|------|----------|
| OrderCreated | 待支付提醒 |
| OrderPaid | 支付成功 |
| OrderCancelled / OrderClosed | 订单关闭 |

`Notifier` 默认 `LogNotifier`；联系方式查 user gRPC。

## 7. 错误码

本域 `19001`–`19003`，见 [15-error-codes.md](../15-error-codes.md#11-notification19000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_NOTIFICATION_MYSQL_DSN` | |
| `MINIMALL_NOTIFICATION_NATS_URL` | |
| `MINIMALL_NOTIFICATION_USER_GRPC` | `user-svc:9001` |
| `MINIMALL_NOTIFICATION_PROVIDER` | `log` |

## 9. 依赖与事件

- NATS 消费者；user 查联系方式
- 不在交易强一致路径上

## 10. 相关文档

- [09-async-events.md](../09-async-events.md)
- [order.md](order.md)

---

导航：[上一篇 search](search.md) | [服务索引](README.md) | [返回总索引](../README.md)
