# 11 可观测性

## 1. 目标

一期具备：**结构化日志**、**RED/USE 核心指标**、**请求级 Trace**，能排查下单与支付链路。

## 2. OpenTelemetry

- 每个服务启动时初始化 TracerProvider + MeterProvider
- 导出：OTLP → `otel-collector` → Prometheus / **Tempo** / Loki
- Trace 后端一期默认 **Tempo**（Compose `obs` profile）；若替换为 Jaeger，仍走 OTLP，业务代码不变
- 传播：W3C Trace Context；HTTP 与 gRPC 中间件自动注入
- `trace_id` 写入日志字段

资源属性：`service.name=order-svc`, `deployment.environment=dev`。

## 3. 指标（Prometheus）

### 3.1 通用

| 指标 | 类型 | 说明 |
|------|------|------|
| `http_server_request_duration_seconds` | histogram | 网关 |
| `grpc_server_handling_seconds` | histogram | 各 svc |
| `process_runtime_go_*` | gauge | Go runtime |
| `db_query_duration_seconds` | histogram | 可选 |

### 3.2 业务

| 指标 | 说明 |
|------|------|
| `minimall_orders_created_total` | 下单计数 |
| `minimall_orders_paid_total` | 支付成功订单 |
| `minimall_inventory_reserve_fail_total` | 预留失败 |
| `minimall_payment_callbacks_total{result=}` | 回调结果 |
| `minimall_nats_consumer_lag` | 若可得 |

网关暴露 `GET /metrics`；gRPC 服务可附带 HTTP sidecar 端口或独立 `:9100`。

## 4. 日志

- JSON 一行一条；级别 info 默认
- 字段：`ts`, `level`, `msg`, `service`, `trace_id`, `request_id`, `user_id?`, `order_id?`
- 输出 stdout；Promtail → Loki
- 禁止记录敏感信息（见安全文档）

## 5. Tracing 关键跨度

下单示例 span 树：

```text
gateway POST /orders
  └─ order.CreateOrder
       ├─ user.GetAddress
       ├─ product.BatchGetSkus
       ├─ coupon.LockCoupon
       ├─ inventory.ReserveStock
       ├─ db.InsertOrder
       └─ payment.CreatePayment
```

支付回调：

```text
payment.Callback
  └─ nats.Publish PaymentSucceeded
order.ConsumePaymentSucceeded
  ├─ db.MarkPaid
  └─ inventory.CommitStock
```

## 6. Grafana 仪表盘（建议）

1. 概览：QPS、错误率、P95 延迟（按服务）
2. 交易：下单/支付成功率、库存失败率
3. 依赖：MySQL / Redis / NATS 可用性

## 7. 告警基线（演示可先手动）

| 条件 | 级别 |
|------|------|
| 网关 5xx 率 > 5% 持续 5m | critical |
| order 下单 P95 > 2s 持续 10m | warning |
| payment 回调验签失败突增 | critical |
| NATS 消费者持续未 ack | warning |
| MySQL 连接耗尽 | critical |

## 8. 健康检查

| 端点 | 行为 |
|------|------|
| `/healthz` | 进程活着 |
| `/readyz` | MySQL/NATS 等依赖可达 |

编排与探针只用 ready 做流量切换。

## 9. 相关文档

- [10-deployment.md](10-deployment.md)
- [03-tech-stack.md](03-tech-stack.md)
- [12-security.md](12-security.md)

---

导航：[上一篇 10 部署](10-deployment.md) | [下一篇 12 安全](12-security.md) | [返回索引](README.md)
