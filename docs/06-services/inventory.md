# Inventory Service

## 1. 职责与非职责

**职责**：SKU 可用库存、预留（Reserve）、确认（Commit）、释放（Release）、超时释放。

**非职责**：创建订单、支付、商品主数据。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `inventory-svc` |
| gRPC | `9008` |
| DB | `minimall_inventory` |
| 经 Gateway | 默认否；可选只读 `GET /api/v1/inventories/{sku_id}` |

详见 [14-port-matrix.md](../14-port-matrix.md)。主调用方为 **order**（内网 gRPC）。

## 3. 对外 REST（可选）

`GET /api/v1/inventories/{sku_id}` → `{ "sku_id", "available" }`

## 4. 对内 gRPC（主契约）

```protobuf
service InventoryService {
  rpc InitStock(InitStockRequest) returns (google.protobuf.Empty);
  rpc GetStock(GetStockRequest) returns (Stock);
  rpc ReserveStock(ReserveStockRequest) returns (ReserveStockResponse);
  rpc CommitStock(CommitStockRequest) returns (google.protobuf.Empty);
  rpc ReleaseStock(ReleaseStockRequest) returns (google.protobuf.Empty);
}
```

- **Reserve**：按 `order_id` 幂等；`available -= qty`，`reserved += qty`；不足 → `13001`
- **Commit**：预留中 → 已确认；`reserved -= qty`
- **Release**：归还 available；已确认不可释放（售后退款二期）

## 5. 数据模型

表：`stocks`、`stock_reservations`、`stock_reservation_items`（可选 `stock_ledger`）。见 [08-data-models.md](../08-data-models.md#4-inventory)。

## 6. 关键流程

超时扫描 `expires_at` 自动 Release；与 order 关单双保险。事件：`StockReserved` / `StockCommitted` / `StockReleased`。

## 7. 错误码

本域 `13001`–`13004`，见 [15-error-codes.md](../15-error-codes.md#5-inventory13000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_INVENTORY_MYSQL_DSN` | |
| `MINIMALL_INVENTORY_GRPC_ADDR` | `:9008` |
| `MINIMALL_INVENTORY_DEFAULT_TTL` | `1800` |
| `MINIMALL_INVENTORY_NATS_URL` | |

## 9. 依赖与事件

- 被 order 同步 Reserve/Commit/Release
- 生产库存事件

## 10. 相关文档

- [order.md](order.md)
- [09-async-events.md](../09-async-events.md)

---

导航：[上一篇 product](product.md) | [下一篇 cart](cart.md) | [服务索引](README.md)
