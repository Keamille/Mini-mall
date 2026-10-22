# internal/user

**user** 的私有实现目录。其他服务不得 import 本目录。

| 目录 | 用途 |
|---|---|
| `app/` | 注册、登录、地址等用例 |
| `domain/` | 用户、地址、ID |
| `port/` | 仓储接口 |
| `adapter/http/` | 路由 |
| `adapter/mysql/` | MySQL |
| `config/` | DSN 与监听地址 |
| `server/` | 组装并启动 |

`cmd/user/main.go` 只调用 `server.Run()`。
