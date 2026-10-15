# cmd/product

Service entrypoint template for **product**.

Run `go run ./cmd/product` (HTTP `:8082`). Acceptance: `powershell -File scripts/accept-product.ps1`. The entrypoint should only load configuration, initialize observability, assemble dependencies, start transports/consumers, and handle graceful shutdown.

Service contract: [docs/06-services/product.md](../../docs/06-services/product.md)
