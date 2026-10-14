# 03 技术栈

本文定义 **一期正式选型**。接口契约、Compose 服务名与端口矩阵均按此栈编写；实现时勿并行引入第二套同类组件。

## 1. 正式选型

| 层 | 选型 | 备注 |
|----|------|------|
| 语言 | Go 1.22+ | Go Modules |
| 仓库形态 | Monorepo | `cmd/` + `internal/` + `api/proto/` |
| 对外 HTTP | Gin | 仅 gateway（及支付回调附属 HTTP） |
| 服务间 RPC | gRPC + Protobuf | proto 包 `minimall.<svc>.v1` |
| 配置 | Viper + 环境变量 | 密钥只走 env / secret |
| 关系库 | MySQL 8 | 每服务独立 database |
| SQL 访问 | sqlx 或 `database/sql` | 不用 GORM 作为主路径 |
| 迁移 | golang-migrate | 版本化 SQL |
| 缓存 | Redis 7 | 限流、会话辅助；购物车一期用 MySQL |
| 消息 | NATS JetStream | 领域事件、至少一次投递 |
| 搜索 | OpenSearch 2.x | 索引由 search-svc 维护 |
| 鉴权 | JWT Access + Refresh Token | Access 短过期 |
| 可观测 | OpenTelemetry + Prometheus + Grafana + Loki | Trace 后端默认 **Tempo**（obs profile；也可用 Jaeger 替换镜像，契约仍为 OTLP） |
| 部署 | Docker Compose | 开发/演示主路径；K8s 不在一期 |
| 容器镜像 | `golang:1.22-bookworm` 构建 + `gcr.io/distroless/static` 或 `alpine` 运行 | 多阶段构建 |

## 2. 选型理由（摘要）

- **Gin**：中间件成熟，适合网关。
- **gRPC**：服务间强类型契约，超时与生成代码清晰。
- **MySQL 8**：资料与运维常见；每服务分库。
- **NATS JetStream**：轻量持久化与消费者组，适合本项目事件量。
- **OpenSearch**：商品检索与过滤；与 catalog 事件同步。

## 3. 版本 pin

在 `go.mod` 与 Compose 中 pin 主版本；升级走变更记录。

| 组件 | 版本 |
|------|------|
| Go | 1.22.x |
| MySQL 镜像 | `mysql:8.4` |
| Redis | `redis:7-alpine` |
| NATS | `nats:2.10`（开启 JetStream） |
| OpenSearch | `opensearchproject/opensearch:2` |
| Tempo | Compose `obs` profile 默认镜像 |

## 4. 第三方支付 / 短信

- 接口：`PaymentProvider` / `Notifier`
- 一期默认：`MockProvider` / `LogNotifier`
- 真实渠道：独立 adapter，密钥仅环境变量；不绑定某一家 SDK 为必选依赖

## 5. 二期再议（非一期选项）

以下不作为一期并行方案，仅作后续演进备忘：PostgreSQL、Kafka、K8s、真实支付/短信渠道、Meilisearch 降级演示。

## 6. 相关文档

- 目录结构：[04-project-structure.md](04-project-structure.md)
- 端口矩阵：[14-port-matrix.md](14-port-matrix.md)
- 部署：[10-deployment.md](10-deployment.md)

---

导航：[上一篇 02 架构](02-architecture.md) | [下一篇 04 目录结构](04-project-structure.md) | [返回索引](README.md)
