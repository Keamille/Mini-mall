# 06 服务规格索引

各微服务一服务一文。表结构细节以 [08 数据模型](../08-data-models.md) 为准；端口以 [14 端口矩阵](../14-port-matrix.md) 为准；错误码以 [15 错误码](../15-error-codes.md) 为准。

## 阅读顺序（交易链路）

1. [gateway.md](gateway.md) — 接入与路由  
2. [user.md](user.md) → [product.md](product.md)  
3. [inventory.md](inventory.md) → [cart.md](cart.md)  
4. [order.md](order.md) → [payment.md](payment.md) → [coupon.md](coupon.md)  
5. [search.md](search.md) → [notification.md](notification.md)

## 服务一览

| 文档 | Compose | gRPC | 对外 |
|------|---------|------|------|
| [gateway](gateway.md) | `gateway` | — | HTTP `:8080` |
| [user](user.md) | `user-svc` | 9001 | 经网关 |
| [product](product.md) | `product-svc` | 9002 | 经网关 |
| [inventory](inventory.md) | `inventory-svc` | 9008 | 内网为主 |
| [cart](cart.md) | `cart-svc` | 9003 | 经网关 |
| [order](order.md) | `order-svc` | 9004 | 经网关 |
| [payment](payment.md) | `payment-svc` | 9005 | 经网关 |
| [coupon](coupon.md) | `coupon-svc` | 9006 | 经网关 |
| [search](search.md) | `search-svc` | 9007 | 经网关 |
| [notification](notification.md) | `notification-svc` | 9009 | 可选 |

## 统一章节模板

每篇服务文档按以下顺序组织：

1. 职责与非职责  
2. Docker / 端口（一行 + 链到 14）  
3. 对外 REST（经网关）  
4. 对内 gRPC  
5. 数据模型（要点 + 链到 08）  
6. 关键流程（若有）  
7. 错误码（本域 + 链到 15）  
8. 配置项  
9. 依赖与事件  
10. 相关文档 + 导航

---

导航：[上一篇 05 编码规范](../05-coding-standards.md) | [下一篇 gateway](gateway.md) | [返回总索引](../README.md)
