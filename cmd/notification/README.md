# cmd/notification

Service entrypoint template for **notification**.

Copy `main.go.tmpl` to `main.go` only when implementing this service. The entrypoint should only load configuration, initialize observability, assemble dependencies, start transports/consumers, and handle graceful shutdown.

Service contract: [docs/06-services/notification.md](../../docs/06-services/notification.md)
