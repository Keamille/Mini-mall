# internal/search

Private implementation area for **search**. No other service may import this directory.

| Directory | Intended contents |
|---|---|
| `app/` | Application use cases and orchestration |
| `domain/` | Entities, value objects, domain errors |
| `port/` | Repository, RPC client, publisher interfaces |
| `adapter/` | gRPC/HTTP, MySQL and NATS implementations |
| `config/` | Service configuration mapping |
| `server/` | Dependency assembly and lifecycle wiring |

Do not add business logic until the relevant proto, migration, and service contract are agreed.
