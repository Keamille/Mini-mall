# cmd/notification

**notification** 服务的启动入口模板。

实现本服务时，将 `main.go.tmpl` 复制为 `main.go`。入口只负责：加载配置、初始化可观测性、组装依赖、启动传输层/消费者，以及优雅退出。

服务规格：[docs/06-services/notification.md](../../docs/06-services/notification.md)