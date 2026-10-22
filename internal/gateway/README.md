# internal/gateway

Private implementation area for **gateway**. No other service may import this directory.

| Directory | Intended contents |
|---|---|
| `app/` | Application use cases and orchestration |
| `domain/` | Entities, value objects, domain errors |
| `port/` | Repository, RPC client, publisher interfaces |
| `adapter/` | gRPC/HTTP, MySQL and NATS implementations |
| `config/` | Service configuration mapping |
| `server/` | Dependency assembly and lifecycle wiring |

Service code lives in these directories. `cmd/<service>/main.go` only calls `server.Run()`.
