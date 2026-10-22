# internal/search

**search** 的私有实现目录。其他服务不得 import 本目录。

| 目录 | 用途 |
|---|---|
| `app/` | 应用用例与编排 |
| `domain/` | 实体、值对象、领域错误 |
| `port/` | 仓储、RPC 客户端、发布者等出站接口 |
| `adapter/` | gRPC/HTTP、MySQL、NATS 等适配实现 |
| `config/` | 服务配置映射 |
| `server/` | 依赖组装与生命周期 |

业务代码已在这些目录里。检索引擎在 `adapter/opensearch`（别名 `minimall_products`），没有业务 MySQL。`cmd/search/main.go` 只调用 `server.Run()`。