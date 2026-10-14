# 07 API 约定

## 1. 总则

- 客户端只调用 Gateway 的 **HTTP/JSON**：`https://{host}/api/v1/...`
- 服务间默认 **gRPC**（protobuf）；Gateway 将 HTTP 转为 gRPC
- 版本：URL 与 proto package 均带 `v1`；破坏性变更升 `v2`
- 字符编码：UTF-8；JSON 字段 `snake_case`

## 2. 通用请求头

| Header | 必填 | 说明 |
|--------|------|------|
| `Authorization` | 需登录接口 | `Bearer <access_token>` |
| `X-Request-Id` | 建议 | 客户端生成 UUID；网关缺失时补全 |
| `Idempotency-Key` | 写关键写 | 下单、支付、领券等；TTL ≥ 24h |
| `Accept-Language` | 可选 | 一期可忽略，预留 |

## 3. 统一成功响应

列表与对象可直接返回资源；若需包一层：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "..."
}
```

一期推荐：**成功时直接返回资源 JSON**；错误时返回统一错误体（减少套娃）。网关与文档示例保持一致即可。

## 4. 统一错误体

```json
{
  "code": 10001,
  "message": "invalid argument: sku_id required",
  "details": [
    { "field": "sku_id", "reason": "required" }
  ],
  "request_id": "a1b2c3d4"
}
```

- `code`：业务码，见 [15-error-codes.md](15-error-codes.md)
- HTTP 状态码与业务码同时返回，客户端优先看 `code`

## 5. 分页

### 请求

| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `page` | int | 1 | 从 1 开始 |
| `page_size` | int | 20 | 最大 100 |

游标分页（大数据量列表，如订单）可选：

| 参数 | 说明 |
|------|------|
| `cursor` | 上一页返回的 `next_cursor` |
| `limit` | 同 page_size |

### 响应

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 100,
  "next_cursor": ""
}
```

## 6. 鉴权

- Access Token：JWT，建议过期 15–30 分钟；claims 至少含 `sub`(user_id)、`exp`、`iat`、`jti`
- Refresh Token：不透明随机串或 JWT，存 Redis/DB，可撤销
- 网关校验 Access；后端 gRPC 可通过 metadata 传递 `x-user-id`（仅内网信任网关）
- 公开接口：注册、登录、商品详情、搜索
- 用户接口：购物车、下单、地址等（需 `sub`）
- 管理接口（一期可简化）：商品写操作可用独立 `admin` 角色 claim 或内网 Basic（文档标注，勿写死密钥）

## 7. REST 资源命名

与 [gateway 路由表](06-services/gateway.md#3-对外-rest路由表) 对齐：

```text
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
GET    /api/v1/users/me
PATCH  /api/v1/users/me
*      /api/v1/users/me/addresses*

GET    /api/v1/categories
GET    /api/v1/products
GET    /api/v1/products/{id}
POST   /api/v1/products                    # admin
PATCH  /api/v1/products/{id}               # admin
POST   /api/v1/products/{id}/skus          # admin
PATCH  /api/v1/skus/{id}                   # admin

GET    /api/v1/cart
PUT    /api/v1/cart/items
DELETE /api/v1/cart/items/{sku_id}
PATCH  /api/v1/cart/items/{sku_id}
POST   /api/v1/cart/clear

POST   /api/v1/orders
GET    /api/v1/orders
GET    /api/v1/orders/{id}
POST   /api/v1/orders/{id}/cancel

POST   /api/v1/payments/{id}/mock-confirm  # 仅开发/演示
POST   /api/v1/payments/callbacks/{provider}

GET    /api/v1/coupons/templates
POST   /api/v1/coupons/templates/{id}/claim
GET    /api/v1/coupons/mine
POST   /api/v1/coupons/quote

GET    /api/v1/search/products?q=

GET    /api/v1/inventories/{sku_id}        # 可选只读
GET    /api/v1/notifications/mine          # 可选
POST   /api/v1/notifications/{id}/read     # 可选
```

详细字段见 [06-services/](06-services/)。

## 8. gRPC 约定

- Package：`minimall.<service>.v1`
- 服务名：`UserService`、`OrderService` 等
- 超时：网关→服务默认 3s；下单编排可 5–10s；禁止无超时
- 重试：仅对幂等读或明确幂等写；下单默认不自动重试（靠 Idempotency-Key）
- 健康检查：标准 `grpc.health.v1.Health`

## 9. 幂等

- 客户端传 `Idempotency-Key`
- 服务端以 `(user_id, key)` 或全局 `key` 存结果指纹
- 相同 key + 相同请求体：返回首次结果
- 相同 key + 不同请求体：`409` / code `10005`

## 10. 相关文档

- 服务接口：[06-services/](06-services/)
- 错误码：[15-error-codes.md](15-error-codes.md)
- 安全：[12-security.md](12-security.md)

---

导航：[上一篇 服务规格](06-services/README.md) | [下一篇 08 数据模型](08-data-models.md) | [返回索引](README.md)
