# Search Service

## 1. 职责与非职责

**职责**：商品文档索引、关键词/过滤检索、消费商品变更事件。

**非职责**：商品主数据源、下单、库存真源。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `search-svc` |
| gRPC | `9007` |
| 引擎 | OpenSearch `opensearch:9200` |

详见 [14-port-matrix.md](../14-port-matrix.md)。一期搜索引擎为 **OpenSearch 2.x**（已锁定）。

## 3. 对外 REST

`GET /api/v1/search/products` — Query：`q`、`category_id`、`price_min`/`price_max`、`page`/`page_size`、`sort`

## 4. 对内 gRPC

```protobuf
service SearchService {
  rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);
  rpc ReindexProduct(ReindexProductRequest) returns (google.protobuf.Empty);
}
```

索引别名：`minimall_products`（物理名 `minimall_products_v1`）。

## 5. 数据模型

主数据在 OpenSearch 文档（`product_id`、`title`、`min_price_cent`、`status` 等）；无独立业务 MySQL 库。见服务内索引字段约定（原规格）。

## 6. 关键流程

消费 `ProductUpserted` / `ProductTakenDown` → upsert/删文档；`event_id` 幂等。

## 7. 错误码

本域 `18001`–`18002`，见 [15-error-codes.md](../15-error-codes.md#10-search18000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_SEARCH_OPENSEARCH_URL` | `http://opensearch:9200` |
| `MINIMALL_SEARCH_GRPC_ADDR` | `:9007` |
| `MINIMALL_SEARCH_NATS_URL` | |
| `MINIMALL_SEARCH_INDEX` | `minimall_products` |

## 9. 依赖与事件

- OpenSearch、NATS；可选 product 全量重建

## 10. 相关文档

- [product.md](product.md)
- [09-async-events.md](../09-async-events.md)

---

导航：[上一篇 coupon](coupon.md) | [下一篇 notification](notification.md) | [服务索引](README.md)
