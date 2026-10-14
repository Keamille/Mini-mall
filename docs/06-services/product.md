# Product Service

## 1. 职责与非职责

**职责**：类目、SPU/SKU、上下架、商品详情读模型。

**非职责**：可售库存（inventory）、搜索索引维护（发事件由 search 消费）、购物车。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `product-svc` |
| gRPC | `9002` |
| DB | `minimall_product` |

详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

- `GET /api/v1/products` — 分页；默认仅上架
- `GET /api/v1/products/{id}` — 含 skus（价格等；不含库存真源）
- `GET /api/v1/categories`
- Admin：`POST/PATCH /api/v1/products*`、`POST /api/v1/products/{id}/skus`、`PATCH /api/v1/skus/{id}`

变更后 outbox：`ProductUpserted` / `ProductTakenDown`。

## 4. 对内 gRPC

```protobuf
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product);
  rpc GetSkus(GetSkusRequest) returns (GetSkusResponse);
  rpc BatchGetSkus(BatchGetSkusRequest) returns (BatchGetSkusResponse);
}
```

`BatchGetSkus` 供 cart/order；**下单价格以快照写入订单**。

## 5. 数据模型

表：`categories`、`products`、`skus`。见 [08-data-models.md](../08-data-models.md#3-product)。  
SKU 新建后需初始化 inventory（Admin 流程调 `InitStock` 或脚本）。

## 6. 关键流程

上架变更 → outbox → NATS → search 索引更新。

## 7. 错误码

本域 `12001`–`12005`，见 [15-error-codes.md](../15-error-codes.md#4-product12000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_PRODUCT_MYSQL_DSN` | |
| `MINIMALL_PRODUCT_GRPC_ADDR` | `:9002` |
| `MINIMALL_PRODUCT_NATS_URL` | |

## 9. 依赖与事件

- MySQL、NATS（发布）
- 被：gateway、cart、order、search

## 10. 相关文档

- [09-async-events.md](../09-async-events.md)
- [search.md](search.md)

---

导航：[上一篇 user](user.md) | [下一篇 inventory](inventory.md) | [服务索引](README.md)
