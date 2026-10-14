# 04 项目结构（Monorepo）

## 1. 目标布局

```text
Go Minimall/
├── api/
│   └── proto/                 # 对外/对内 protobuf
│       ├── user/v1/
│       ├── product/v1/
│       ├── inventory/v1/
│       ├── cart/v1/
│       ├── order/v1/
│       ├── payment/v1/
│       ├── coupon/v1/
│       ├── search/v1/
│       └── common/v1/         # 分页、空响应等公共类型
├── cmd/
│   ├── gateway/
│   ├── user/
│   ├── product/
│   ├── inventory/
│   ├── cart/
│   ├── order/
│   ├── payment/
│   ├── coupon/
│   ├── search/
│   └── notification/
├── internal/                  # 不对外 import
│   ├── gateway/
│   ├── user/
│   ├── product/
│   ├── inventory/
│   ├── cart/
│   ├── order/
│   ├── payment/
│   ├── coupon/
│   ├── search/
│   └── notification/
├── pkg/                       # 可跨服务复用的轻量库
│   ├── auth/                  # JWT 解析（无业务）
│   ├── errx/                  # 统一错误类型
│   ├── middleware/
│   ├── otelx/
│   └── config/
├── deployments/
│   ├── docker-compose.yml
│   ├── docker-compose.obs.yml # 可观测 profile（可选拆分）
│   ├── .env.example
│   └── prometheus/
├── migrations/                # 或放在各服务 internal/*/migrations
│   ├── user/
│   ├── product/
│   └── ...
├── docs/                      # 本开发文档
├── scripts/
│   ├── gen-proto.sh
│   └── migrate.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 2. 单服务内部布局（以 order 为例）

```text
internal/order/
├── app/           # 应用服务 / 用例编排
├── domain/        # 实体、值对象、领域错误
├── port/          # 出站端口接口（repo、rpc、publisher）
├── adapter/
│   ├── http/      # 若该服务也暴露少量 HTTP（一般不推荐，优先网关）
│   ├── grpc/
│   ├── mysql/
│   └── nats/
├── config/
└── server/        # 组装 DI、启停
```

`cmd/order/main.go` 仅负责：加载配置、初始化可观测、启动 gRPC/消费者、优雅退出。

## 3. 依赖规则

| 允许 | 禁止 |
|------|------|
| `cmd/*` → `internal/<svc>` | `internal/user` import `internal/order` |
| `internal/<svc>` → `pkg/*`、生成的 `api` | 服务间共享对方 MySQL repo |
| 服务间调用经 gRPC client / 事件 | 循环依赖 proto |

跨服务 DTO 只出现在 `api/proto` 与事件 payload 中。

## 4. 配置文件约定

- 默认：环境变量（Compose `environment` / `.env`）
- 可选：`configs/<svc>.yaml` 挂载只读卷，密钥仍走 env
- 统一前缀：`MINIMALL_<SVC>_`，例如 `MINIMALL_ORDER_MYSQL_DSN`

## 5. 代码生成

- Protobuf：`buf generate` 或 `protoc` + `scripts/gen-proto.sh`
- 生成物建议提交或 CI 生成二选一；一期建议**提交生成代码**降低本地门槛

## 6. Makefile 常用目标（建议）

```text
make proto          # 生成 pb
make migrate-up     # 全部或按 SVC=
make test
make docker-up      # compose up 基础 + 业务
make docker-down
```

## 7. 相关文档

- 编码规范：[05-coding-standards.md](05-coding-standards.md)
- 端口矩阵：[14-port-matrix.md](14-port-matrix.md)
- 部署：[10-deployment.md](10-deployment.md)

---

导航：[上一篇 03 技术栈](03-tech-stack.md) | [下一篇 05 编码规范](05-coding-standards.md) | [返回索引](README.md)
