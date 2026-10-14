# User Service

## 1. 职责与非职责

**职责**：注册/登录、资料、收货地址、Access/Refresh Token 签发与刷新撤销。

**非职责**：订单、支付、复杂 RBAC（一期仅 `user`/`admin` claim）。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `user-svc` |
| gRPC | `9001` |
| DB | `minimall_user` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

### POST `/api/v1/auth/register`

请求：`email` / `phone`、`password`、`nickname`  
响应 `201`：`user` + `access_token` + `refresh_token` + `expires_in`  
错误：已注册 → `11001`

### POST `/api/v1/auth/login`

凭证登录；错误 → `11002`

### POST `/api/v1/auth/refresh`

请求：`refresh_token`；旋转作废旧 refresh；无效 → `11003`

### GET/PATCH `/api/v1/users/me`

资料读写（无 `password_hash`）

### 地址

- `GET/POST /api/v1/users/me/addresses`
- `PATCH/DELETE /api/v1/users/me/addresses/{id}`

## 4. 对内 gRPC

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc GetAddress(GetAddressRequest) returns (Address);
  rpc ValidateUserActive(ValidateUserActiveRequest) returns (google.protobuf.Empty);
}
```

## 5. 数据模型

表：`users`、`user_addresses`、`refresh_tokens`。字段见 [08-data-models.md](../08-data-models.md#2-user)。

## 6. 关键流程：登录

查用户 → 校验密码哈希 → 状态正常 → 签发 Access → 存储 Refresh hash → 返回双 token。

## 7. 错误码

本域 `11001`–`11005`，总表见 [15-error-codes.md](../15-error-codes.md#3-user11000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_USER_MYSQL_DSN` | |
| `MINIMALL_USER_JWT_*` | 与网关验签配对 |
| `MINIMALL_USER_ACCESS_TTL` | 默认 `30m` |
| `MINIMALL_USER_REFRESH_TTL` | 默认 `720h` |
| `MINIMALL_USER_GRPC_ADDR` | `:9001` |

## 9. 依赖与事件

- MySQL
- 一期可不发领域事件

## 10. 相关文档

- [12-security.md](../12-security.md)
- [15-error-codes.md](../15-error-codes.md)

---

导航：[上一篇 gateway](gateway.md) | [下一篇 product](product.md) | [服务索引](README.md)
