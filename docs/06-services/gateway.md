# Gateway（API 网关）

## 1. 职责与非职责

**职责**

- 对外暴露 HTTP `/api/v1/*`
- JWT Access Token 校验与公开路由白名单
- 路由到后端 gRPC
- 限流、请求日志、`X-Request-Id` / trace 注入
- 统一错误体映射（gRPC status → HTTP + code）

**非职责**

- 不持有业务数据库
- 不做库存、计价、支付状态机
- **不直连 `inventory-svc`**（预留/确认由 order 内网调用）
- 不做重业务聚合

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `gateway` |
| HTTP | `8080`（`localhost:8080`） |

完整矩阵见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST（路由表）

| 方法 | 路径 | 目标 | 鉴权 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | user | 否 |
| POST | `/api/v1/auth/login` | user | 否 |
| POST | `/api/v1/auth/refresh` | user | 否 |
| GET/PATCH | `/api/v1/users/me` | user | 是 |
| * | `/api/v1/users/me/addresses*` | user | 是 |
| GET | `/api/v1/products*` | product | 否 |
| GET | `/api/v1/categories` | product | 否 |
| POST/PATCH | `/api/v1/products*`、`/api/v1/skus*` | product | Admin |
| * | `/api/v1/cart*` | cart | 是 |
| * | `/api/v1/orders*` | order | 是 |
| POST | `/api/v1/payments/callbacks/*` | payment | 签名校验 |
| POST | `/api/v1/payments/{id}/mock-confirm` | payment | 是（仅 dev） |
| GET | `/api/v1/coupons/templates` | coupon | 否 |
| POST | `/api/v1/coupons/templates/{id}/claim` | coupon | 是 |
| GET | `/api/v1/coupons/mine` | coupon | 是 |
| POST | `/api/v1/coupons/quote` | coupon | 是 |
| GET | `/api/v1/search/products` | search | 否 |
| GET | `/api/v1/inventories/{sku_id}` | inventory | 否（可选只读） |
| GET | `/api/v1/notifications/mine` | notification | 是（可选） |
| GET | `/healthz` | 本地 | 否 |

路径约定与 [07-api-conventions.md](../07-api-conventions.md) 对齐。

## 4. 对内 gRPC

网关作为 gRPC **客户端**，下游见 [14 §3](../14-port-matrix.md#3-gateway-下游地址约定)。不包含 inventory / notification 主路径（可选路由除外）。

## 5. 数据模型

无业务库。

## 6. 关键流程

中间件顺序：Recover → RequestID → AccessLog → OTel/Metrics → CORS → RateLimit → Auth → Handler。

## 7. 错误码

透传下游业务码；无业务码时按 [15 通用码](../15-error-codes.md#2-通用10000) 与 gRPC 映射。

| gRPC | HTTP | 默认 code |
|------|------|-----------|
| InvalidArgument | 400 | 10001 |
| Unauthenticated | 401 | 10002 |
| PermissionDenied | 403 | 10003 |
| NotFound | 404 | 10004 |
| Aborted / AlreadyExists | 409 | 10005 |
| ResourceExhausted | 429 | 10006 |
| Internal/Unknown | 500 | 10007 |

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_GATEWAY_HTTP_ADDR` | 默认 `:8080` |
| `MINIMALL_GATEWAY_JWT_*` | 验签（与 user 签发配对） |
| `MINIMALL_GATEWAY_*_GRPC` | user/product/cart/order/payment/coupon/search 地址 |
| `MINIMALL_GATEWAY_INVENTORY_GRPC` | 仅当开启库存只读路由 |
| `MINIMALL_GATEWAY_REDIS_ADDR` | 限流 |
| `MINIMALL_GATEWAY_ENV` | `dev`/`prod` |

## 9. 依赖与事件

- 依赖：Redis；上述 gRPC 服务
- 不直接依赖 NATS

## 10. 相关文档

- [07-api-conventions.md](../07-api-conventions.md)
- [12-security.md](../12-security.md)
- [14-port-matrix.md](../14-port-matrix.md)

---

导航：[服务索引](README.md) | [下一篇 user](user.md) | [返回总索引](../README.md)
