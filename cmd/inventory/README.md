# cmd/inventory

**inventory** 服务的启动入口模板。

`main.go` 只调用对应 `internal/<服务>/server.Run()`，由 `server` 加载配置、连接依赖并启动 HTTP/gRPC。

服务规格：[docs/06-services/inventory.md](../../docs/06-services/inventory.md)