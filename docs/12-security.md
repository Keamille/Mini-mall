# 12 安全

## 1. 认证与授权

- 用户认证：Access JWT + Refresh Token 旋转
- 网关统一验签；下游信任 `x-user-id` **仅限内网**
- 服务间：一期 Compose 内网隔离；生产建议 mTLS 或 mesh，或服务短时 JWT
- Admin 商品写接口：JWT `role=admin` 或独立管理端口（不上公网）

## 2. 密钥管理

| 密钥 | 存放 |
|------|------|
| JWT 密钥 / 私钥 | 环境变量 / Docker secret |
| MySQL 密码 | `.env`（不入库） |
| 支付渠道 secret | 环境变量；按 provider 前缀 |
| OpenSearch | 本地可关安全插件；生产开 TLS+账号 |

轮换：Refresh 可批量撤销；JWT 短过期降低泄漏窗口。

## 3. 传输与暴露面

- 本地：HTTP；生产：网关前 HTTPS
- 仅映射必要端口：默认对外 `8080`；DB/Redis/NATS/gRPC **不对公网**
- 支付回调 URL 使用 HTTPS 与验签

## 4. 输入与注入防护

- 所有外部输入校验（长度、枚举、正整数数量）
- SQL 一律参数化绑定
- 日志与错误信息避免回显内部 SQL/堆栈给客户端

## 5. 幂等与重放

- 下单/支付/领券：`Idempotency-Key`
- 支付回调：渠道单号唯一 + 验签 + 时间窗（若渠道支持）
- Refresh Token：一用一旋，重放检测可强制全下线

## 6. 限流与滥用

- 网关 IP / 用户限流（见 gateway 文档）
- 登录与注册更严；可加验证码钩子（二期）
- 库存与券领取防刷：用户维度限流 + 业务校验

## 7. 敏感数据

| 数据 | 要求 |
|------|------|
| 密码 | argon2id/bcrypt；不可逆 |
| Refresh Token | 仅存 hash |
| 支付回调原文 | 按需短期存；脱敏日志 |
| 手机/邮箱 | 最小化展示；通知模板脱敏 |

GDPR/个保法合规细节二期补充；一期至少做到可删除用户的扩展点。

## 8. 依赖与镜像

- 基础镜像固定 digest 或次版本
- `go mod` 定期漏洞扫描（govulncheck）
- 容器非 root 用户运行（distroless/alpine + USER）

## 9. 安全检查清单（上线前）

- [ ] 默认密码与 `dev-only` JWT 密钥已替换
- [ ] Mock 支付确认接口在生产关闭
- [ ] 调试端口与 pprof 不暴露公网
- [ ] CORS 白名单收紧
- [ ] 备份与恢复演练（MySQL 卷）

## 10. 相关文档

- [07-api-conventions.md](07-api-conventions.md)
- [06-services/gateway.md](06-services/gateway.md)
- [06-services/payment.md](06-services/payment.md)
- [14-port-matrix.md](14-port-matrix.md)（暴露面）

---

导航：[上一篇 11 可观测](11-observability.md) | [下一篇 13 路线图](13-roadmap.md) | [返回索引](README.md)
