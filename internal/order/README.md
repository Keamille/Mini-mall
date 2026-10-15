# internal/order

**order** 的私有实现目录。其他服务不得 import 本目录。

| 目录 | 用途 |
|---|---|
| `app/` | 应用用例与编排 |
| `domain/` | 实体、值对象、领域错误 |
| `port/` | 仓储、RPC 客户端、发布者等出站接口 |
| `adapter/` | gRPC/HTTP、MySQL、NATS 等适配实现 |
| `config/` | 服务配置映射 |
| `server/` | 依赖组装与生命周期 |

在对应 proto、迁移与服务规格对齐前，不要写入业务逻辑。