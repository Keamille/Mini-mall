# Go Minimall

Go 电商微服务。浏览器只访问网关；服务之间目前用 HTTP 调用。设计文档里的 gRPC、NATS、OpenSearch 还没换成正式实现。

完整设计见 [docs/README.md](docs/README.md)。

## 本地依赖

MySQL 8.4，由 Compose 提供。账号 `minimall` / `123456`，端口 `3306`。

```powershell
docker compose -f deployments/docker-compose.yml up -d mysql
```

初始化 SQL 只在数据卷为空时执行。已有数据卷要补新表时，见 [deployments/mysql/init/README.md](deployments/mysql/init/README.md)。

## 服务

| 服务 | 端口 | 库 | 作用 |
|---|---|---|---|
| gateway | 8080 | — | 对外入口，校验 JWT 后反向代理 |
| user | 8081 | minimall_user | 注册、登录、地址 |
| product | 8082 | minimall_product | 类目、商品、SKU；写入后同步搜索文档 |
| cart | 8083 | minimall_cart | 购物车 |
| search | 8084 | minimall_search | 上架商品检索（MySQL 文档，不是 OpenSearch） |
| payment | 8085 | minimall_payment | 创建支付单、mock 确认，并通知订单已支付 |
| coupon | 8086 | minimall_coupon | 券模板、领取、试算；锁定 / 解锁 / 核销只给订单直连 |
| notification | 8087 | minimall_notification | 站内信。订单直连写入，同一事件号不重复插入 |
| inventory | 8088 | minimall_inventory | 库存预留、释放、确认 |
| order | 8089 | minimall_order | 下单、取消、标记已支付 |

JWT 开发密钥是 `dev-only-change-me`。金额单位是分。

## 已通的主流程

1. 注册登录后，经网关访问商品、购物车、库存、订单、支付、优惠券、搜索和站内信。
2. 下单带 `Idempotency-Key`。可选 `user_coupon_id`：先锁券，再预留库存，应付金额为商品额减去减免。
3. 取消待支付订单会归还库存并解锁券。
4. mock 确认支付后，订单变为已支付，库存 Commit，券核销。
5. 下单、取消、支付成功各写一条站内信。
6. 创建商品或 SKU 后，搜索索引会更新。搜索服务没启动时，商品写入仍然成功。

内部接口不经网关：`POST /api/v1/orders/:id/paid`、优惠券的 lock / unlock / confirm、搜索文档写入、通知事件写入。

## 还没做

服务间 gRPC、NATS、OpenSearch、超时关单、从购物车下单、真实支付渠道。
