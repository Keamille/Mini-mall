# cmd/gateway

**gateway** 入口：HTTP `:8080`，反代 user / product 并做 JWT 鉴权。

```bash
go run ./cmd/gateway
```

验收：`powershell -File scripts/accept-gateway.ps1`（需 user、product 已启动）。

规格：[docs/06-services/gateway.md](../../docs/06-services/gateway.md)
