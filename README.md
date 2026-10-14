# Go Minimall

Go 电商微服务项目骨架。当前仅包含目录、配置与接口模板；不包含业务实现、数据库迁移或可运行服务。

## 已锁定技术栈

- Go 1.22+
- Gin（Gateway）
- gRPC + Protobuf（服务间通信）
- MySQL 8、Redis 7、NATS JetStream、OpenSearch 2.x
- Docker Compose

完整设计见 [docs/README.md](docs/README.md)。

## 目录

| 路径 | 用途 |
|---|---|
| `api/proto/` | Protobuf 契约模板 |
| `cmd/` | 每个服务的启动入口模板 |
| `internal/` | 服务私有分层模板 |
| `pkg/` | 跨服务轻量公共包模板 |
| `configs/` | 非敏感配置示例 |
| `migrations/` | SQL 迁移占位目录 |
| `deployments/` | Compose 与基础设施模板 |
| `scripts/` | 本地开发脚本模板 |
| `docs/` | 架构与开发规范 |

## 初始化前必做

1. 将 `go.mod` 中的模块路径替换为实际 Git 仓库路径。
2. 依据 `deployments/.env.example` 创建本地 `.env`，不要提交真实密钥。
3. 按 [13 路线图](docs/13-roadmap.md) 从 M0 开始实现。

## 当前状态

所有服务目录只提供 README、`.gitkeep` 与 `main.go.tmpl`。将模板复制为 `main.go` 并补全依赖组装后，才进入可运行实现阶段。
