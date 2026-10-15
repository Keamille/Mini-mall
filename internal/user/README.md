# internal/user

**user** 的私有实现目录。其他服务不得 import 本目录。

当前为同包多文件（阶段 A），尚未拆到 app/domain/adapter。

| 文件 | 用途 |
|---|---|
| `handler.go` | Handler 结构体与 NewHandler |
| `register.go` | 注册 HTTP |
| `login.go` | 登录 HTTP |
| `me.go` | 当前用户资料 |
| `auth_jwt.go` | Access Token 签发与鉴权中间件 |
| `auth_refresh.go` | Refresh Token 签发、旋转与刷新接口 |
| `store.go` | MySQL 读写 |
| `id.go` | ULID 生成 |

目标分层（阶段 B，以后再迁）：

| 目录 | 用途 |
|---|---|
| `app/` | 应用用例与编排 |
| `domain/` | 实体、值对象、领域错误 |
| `port/` | 仓储等出站接口 |
| `adapter/` | gRPC/HTTP、MySQL、NATS |
| `config/` | 服务配置映射 |
| `server/` | 依赖组装与生命周期 |
