# Go Minimall — 电商微服务开发文档

面向生产落地规格的 Go 电商微服务系统设计与开发指南。部署主路径为 **Docker / Docker Compose**；技术栈见 [03 技术栈](03-tech-stack.md)（**一期已锁定**）。

## 阅读路径

| 阶段 | 文档 | 说明 |
|------|------|------|
| 入门 | [01 总览](01-overview.md) → [02 架构](02-architecture.md) → [03 技术栈](03-tech-stack.md) | 边界与正式选型 |
| 工程约定 | [04 目录结构](04-project-structure.md) → [05 编码规范](05-coding-standards.md) → [07 API 约定](07-api-conventions.md) | 开工前对齐 |
| 速查附录 | [14 端口矩阵](14-port-matrix.md) → [15 错误码](15-error-codes.md) | 端口与错误码单一事实源 |
| 领域规格 | [08 数据模型](08-data-models.md) → [09 异步事件](09-async-events.md) → [06 各服务](06-services/) | 按交易链路实现 |
| 上线准备 | [10 部署](10-deployment.md) → [11 可观测](11-observability.md) → [12 安全](12-security.md) → [13 路线图](13-roadmap.md) | Docker 与验收 |

## 文档索引

### 基础

- [01-overview.md](01-overview.md) — 背景、目标、非目标、里程碑
- [02-architecture.md](02-architecture.md) — 整体架构、服务边界、调用关系
- [03-tech-stack.md](03-tech-stack.md) — **一期正式选型**
- [04-project-structure.md](04-project-structure.md) — Monorepo 目录约定
- [05-coding-standards.md](05-coding-standards.md) — 编码、错误码分段、日志、测试规范

### 服务规格

- [06-services/README.md](06-services/README.md) — 服务索引与规格模板
- [gateway](06-services/gateway.md) · [user](06-services/user.md) · [product](06-services/product.md) · [inventory](06-services/inventory.md) · [cart](06-services/cart.md) · [order](06-services/order.md) · [payment](06-services/payment.md) · [coupon](06-services/coupon.md) · [search](06-services/search.md) · [notification](06-services/notification.md)

### 横切与运维

- [07-api-conventions.md](07-api-conventions.md) — REST / gRPC、鉴权、分页、错误体
- [08-data-models.md](08-data-models.md) — 核心表结构与领域模型
- [09-async-events.md](09-async-events.md) — 领域事件、Topic、最终一致性
- [10-deployment.md](10-deployment.md) — Docker Compose 拓扑与环境变量
- [11-observability.md](11-observability.md) — 指标、链路、日志、告警
- [12-security.md](12-security.md) — 鉴权、密钥、限流、幂等
- [13-roadmap.md](13-roadmap.md) — 分阶段交付与验收清单

### 附录（单一事实源）

- [14-port-matrix.md](14-port-matrix.md) — 服务名、端口、依赖、是否经网关暴露
- [15-error-codes.md](15-error-codes.md) — 业务错误码总表

## 核心服务一览

| 服务 | Docker 名 | gRPC | 职责摘要 |
|------|-----------|------|----------|
| gateway | `gateway` | — | 对外 HTTP、鉴权、路由、限流 |
| user | `user-svc` | 9001 | 注册登录、资料、地址 |
| product | `product-svc` | 9002 | SPU/SKU、类目、上下架 |
| cart | `cart-svc` | 9003 | 购物车 CRUD |
| order | `order-svc` | 9004 | 下单、状态机、取消 |
| payment | `payment-svc` | 9005 | 支付单、回调、退款抽象 |
| coupon | `coupon-svc` | 9006 | 券模板、领取、核销 |
| search | `search-svc` | 9007 | 商品索引与检索 |
| inventory | `inventory-svc` | 9008 | 库存预留、确认、释放 |
| notification | `notification-svc` | 9009 | 短信/邮件/站内信（可模拟） |

完整端口与依赖见 [14-port-matrix.md](14-port-matrix.md)。服务边界见 [02-architecture.md](02-architecture.md#3-服务边界一句话)。

## 约定（已锁定）

- 文档语言：中文
- API 前缀：`/api/v1`
- 服务间 RPC：gRPC + Protobuf
- 消息总线：NATS JetStream
- 关系库：MySQL 8（每服务独立 database）
- 搜索：OpenSearch 2.x
- 部署：Docker Compose

选型细节以 [03-tech-stack.md](03-tech-stack.md) 为准。
