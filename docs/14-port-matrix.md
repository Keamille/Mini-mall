# 14 端口与依赖矩阵

**单一事实源**：各服务文档与 Compose 中的端口、服务名以本文为准；服务文只保留本服务一行摘要并回链此处。

## 1. 业务服务

| Compose 名 | 进程 | HTTP（容器内） | gRPC | 经 Gateway 对外 | 主要依赖 |
|------------|------|----------------|------|-----------------|----------|
| `gateway` | gateway | `8080` | — | 是（唯一公网入口） | redis + 下列可路由 svc |
| `user-svc` | user | `/healthz` 可附带 | `9001` | 是（经 gateway） | mysql |
| `product-svc` | product | 可附带 | `9002` | 是 | mysql, nats |
| `cart-svc` | cart | 可附带 | `9003` | 是 | mysql, product |
| `order-svc` | order | 可附带 | `9004` | 是 | mysql, nats, user, product, inventory, payment, coupon, cart |
| `payment-svc` | payment | `8085`（回调附属） | `9005` | 是（mock/回调经 gateway） | mysql, nats |
| `coupon-svc` | coupon | 可附带 | `9006` | 是 | mysql, nats |
| `search-svc` | search | 可附带 | `9007` | 是 | opensearch, nats |
| `inventory-svc` | inventory | 可附带 | `9008` | **否**（order 等内网 gRPC；可选只读经 gateway） | mysql, nats |
| `notification-svc` | notification | 可附带 | `9009` | 可选（站内信列表） | mysql, nats, user |

宿主机映射（开发默认）：

| 用途 | 映射 |
|------|------|
| API | `localhost:8080` → gateway:8080 |
| 支付回调直连（仅本地调试） | `localhost:8085` → payment-svc:8085 |
| gRPC | **不映射公网**；容器间用服务名 |

## 2. 基础设施

| Compose 名 | 端口（容器） | 宿主机映射（开发） |
|------------|--------------|-------------------|
| `mysql` | 3306 | 3306 |
| `redis` | 6379 | 6379 |
| `nats` | 4222（客户端）, 8222（监控） | 同左 |
| `opensearch` | 9200 | 9200 |
| `prometheus` | 9090 | 9090（obs） |
| `grafana` | 3000 | 3000（obs） |
| `tempo` | 4317 OTLP / 3200 | obs profile |
| `loki` | 3100 | obs |

## 3. Gateway 下游地址（约定）

```text
user-svc:9001
product-svc:9002
cart-svc:9003
order-svc:9004
payment-svc:9005
coupon-svc:9006
search-svc:9007
```

`inventory-svc:9008`、`notification-svc:9009` **不由 gateway 作为主下游**（notification 若暴露站内信可读接口，可另加路由）。

## 4. 数据库名

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

## 5. 相关文档

- 部署：[10-deployment.md](10-deployment.md)
- 服务索引：[06-services/README.md](06-services/README.md)

---

导航：[上一篇 13 路线图](13-roadmap.md) | [下一篇 15 错误码](15-error-codes.md) | [返回索引](README.md)
