# Cart Service

## 1. 职责与非职责

**职责**：登录用户购物车行项目增删改查、选中状态。

**非职责**：锁库存、算运费、落订单。

## 2. Docker / 端口

| 项 | 值 |
|----|-----|
| Compose 名 | `cart-svc` |
| gRPC | `9003` |
| DB | `minimall_cart` |

一期用 **MySQL**（非 Redis）。详见 [14-port-matrix.md](../14-port-matrix.md)。

## 3. 对外 REST

- `GET /api/v1/cart` — 填充商品展示字段（调 product）
- `PUT /api/v1/cart/items` — `{ sku_id, quantity }`；`quantity=0` 删除
- `DELETE /api/v1/cart/items/{sku_id}`
- `PATCH /api/v1/cart/items/{sku_id}` — `checked` / `quantity`
- `POST /api/v1/cart/clear`

## 4. 对内 gRPC

```protobuf
service CartService {
  rpc GetCart(GetCartRequest) returns (Cart);
  rpc ClearCheckedItems(ClearCheckedItemsRequest) returns (google.protobuf.Empty);
}
```

## 5. 数据模型

表：`cart_items`（UNIQUE `user_id,sku_id`）。见 [08-data-models.md](../08-data-models.md#5-cart)。

## 6. 关键流程

读车 → BatchGetSkus → 可选查库存可用性 → 返回。

## 7. 错误码

本域 `14001`–`14003`，见 [15-error-codes.md](../15-error-codes.md#6-cart14000)。

## 8. 配置项

| 变量 | 说明 |
|------|------|
| `MINIMALL_CART_MYSQL_DSN` | |
| `MINIMALL_CART_PRODUCT_GRPC` | `product-svc:9002` |
| `MINIMALL_CART_GRPC_ADDR` | `:9003` |

## 9. 依赖与事件

- product（同步）；inventory 可选只读
- 被 order 可选调用 ClearChecked

## 10. 相关文档

- [product.md](product.md)
- [order.md](order.md)

---

导航：[上一篇 inventory](inventory.md) | [下一篇 order](order.md) | [服务索引](README.md)
