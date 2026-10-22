# Go Minimall

Go 电商微服务。浏览器只访问网关的 HTTP。服务之间原先互相 HTTP 调用的路径已改为 gRPC。订单创建后的站内信在保留 gRPC 的同时，会经 **NATS JetStream** 再投递一条（同一 `event_id` 幂等）。网关限流用 Redis。待支付订单超时后由 order 进程关单（status **40**），并释放库存、解锁券。

完整设计见 [docs/README.md](docs/README.md)。

## 本地依赖

MySQL 8.4、Redis 7、NATS 2.10（JetStream）、OpenSearch 2.19，由 Compose 提供。MySQL 账号 `minimall` / `123456`，端口 `3306`；Redis `6379`；NATS 客户端 `4222`，监控 `8222`；OpenSearch `9200`（单节点，未开安全插件）。

```powershell
docker compose -f deployments/docker-compose.yml up -d mysql redis nats opensearch
```

网关对 `/api/v1/*` 按客户端 IP 限流（默认每 60 秒 120 次）。环境变量：`MINIMALL_GATEWAY_REDIS_ADDR`（默认 `127.0.0.1:6379`）、`MINIMALL_GATEWAY_RATE_LIMIT`、`MINIMALL_GATEWAY_RATE_WINDOW_SEC`。Redis 不可用时放行。设为 `off` 可关闭限流。

订单 / 通知连 `nats://127.0.0.1:4222`。NATS 不可用时进程仍启动：下单站内信仍可走 gRPC；NATS 消费者不启动。

订单超时关单：`MINIMALL_ORDER_PAYMENT_TTL`（秒，默认 `1800`）、`MINIMALL_ORDER_CLOSE_INTERVAL_SEC`（扫描间隔，默认 `5`）。下单写入 `pay_expire_at`；后台将过期且仍为待支付的订单改为 **40**，并 gRPC Release 库存、Unlock 券、写 `order_closed` 站内信。

搜索走 OpenSearch，索引别名 `minimall_products`（物理索引 `minimall_products_v1`）。`MINIMALL_SEARCH_OPENSEARCH_URL` 默认 `http://127.0.0.1:9200`，`MINIMALL_SEARCH_INDEX` 默认 `minimall_products`。检索仍是 `q`、`category_id`、价格区间、`sort`（`new` / `price_asc` / `price_desc`）。商品写入后的 gRPC `UpsertProduct` 不变。搜索进程不再连接 MySQL。

初始化 SQL 只在数据卷为空时执行。已有库缺 `pay_expire_at` 时，order 启动会尝试 `ALTER`；也可执行 `migrations/order/000002_pay_expire.up.sql`。说明见 [deployments/mysql/init/README.md](deployments/mysql/init/README.md)。

## 服务

| 服务 | HTTP | gRPC | 其它 | 作用 |
|---|---|---|---|---|
| gateway | 8080 | — | Redis | 对外入口，JWT、限流、反向代理 |
| user | 8081 | — | — | 注册、登录、地址 |
| product | 8082 | — | — | 类目、商品、SKU；写入后 gRPC 同步搜索 |
| cart | 8083 | — | — | 购物车 |
| search | 8084 | 9007 | OpenSearch | HTTP 检索；gRPC `UpsertProduct` |
| payment | 8085 | — | — | 支付；确认后 gRPC `MarkPaid` |
| coupon | 8086 | 9006 | — | 模板 HTTP；锁券等 gRPC |
| notification | 8087 | 9009 | NATS 消费 | HTTP 读站内信；gRPC `RecordEvent`；JetStream `notify-worker` |
| inventory | 8088 | 9008 | — | 查询 HTTP；预留/释放/确认 gRPC |
| order | 8089 | 9004 | NATS + 关单扫描 | 下单/取消；`MarkPaid`；`order_created` NATS；超时关单 |

JWT 开发密钥是 `dev-only-change-me`。金额单位是分。

Protobuf：`.\scripts\gen-proto.ps1`（需 `protoc` 与 go/grpc 插件）。

## 已通的主流程

1. 注册登录后，经网关访问商品、购物车、库存、订单、支付、优惠券、搜索和站内信。
2. 下单带 `Idempotency-Key`。可选券：gRPC 锁券 → gRPC 预留库存；响应含 `pay_expire_at`。
3. 下单成功后：gRPC 写站内信，并 Publish `minimall.order.order_created`。
4. 取消：手动取消 status **30**；超时未支付：扫描关单 status **40**（释放库存、解锁券、`order_closed` 通知）。
5. mock 支付：gRPC `MarkPaid` → 库存 Commit、券核销；取消与支付的站内信仍仅 gRPC。
6. 商品写入后 gRPC 更新 OpenSearch 索引。

各服务代码在 `internal/<服务>/` 下按 `app`（用例）、`domain`、`port`、`adapter`（http / grpc / mysql / nats，搜索是 opensearch）、`config`、`server` 分目录。`cmd/<服务>/main.go` 只启动 `server.Run()`。

## 还没做

user/cart 的 gRPC、支付成功改纯事件、从购物车下单、真实支付、网关 HTTP→后端 gRPC、outbox、`order_paid` / `order_cancelled` 的 NATS。
