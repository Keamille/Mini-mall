# Go Minimall

Go 电商微服务。浏览器只访问网关的 HTTP。服务之间凡是原来互相调用的路径，都已改成 gRPC。网关对外的接口、各服务自己的 HTTP 仍保留，给浏览器和现有脚本用。

完整设计见 [docs/README.md](docs/README.md)。

## 本地依赖

MySQL 8.4，由 Compose 提供。账号 `minimall` / `123456`，端口 `3306`。

```powershell
docker compose -f deployments/docker-compose.yml up -d mysql
```

初始化 SQL 只在数据卷为空时执行。已有数据卷要补新表时，见 [deployments/mysql/init/README.md](deployments/mysql/init/README.md)。

## 服务

| 服务 | HTTP | gRPC | 库 | 作用 |
|---|---|---|---|---|
| gateway | 8080 | — | — | 对外入口，校验 JWT 后反向代理到各服务 HTTP |
| user | 8081 | — | minimall_user | 注册、登录、地址。没有别的服务调用它 |
| product | 8082 | — | minimall_product | 类目、商品、SKU。写入后经 gRPC 同步搜索 |
| cart | 8083 | — | minimall_cart | 购物车。没有别的服务调用它 |
| search | 8084 | 9007 | minimall_search | 检索仍是 HTTP。商品写入走 gRPC `UpsertProduct` |
| payment | 8085 | — | minimall_payment | 创建支付单、mock 确认。确认后经 gRPC 通知订单 |
| coupon | 8086 | 9006 | minimall_coupon | 模板、领取、试算走 HTTP。锁定 / 解锁 / 核销走 gRPC |
| notification | 8087 | 9009 | minimall_notification | 站内信读取走 HTTP。订单写入走 gRPC `RecordEvent` |
| inventory | 8088 | 9008 | minimall_inventory | 查询和初始化仍是 HTTP。预留、释放、确认走 gRPC |
| order | 8089 | 9004 | minimall_order | 下单、取消走 HTTP。支付服务标记已支付走 gRPC `MarkPaid` |

JWT 开发密钥是 `dev-only-change-me`。金额单位是分。

服务间契约在 `api/proto/*/v1/*.proto`。生成 Go 代码：

```powershell
.\scripts\gen-proto.ps1
```

需要本机已安装 `protoc`、`protoc-gen-go` 和 `protoc-gen-go-grpc`。输出在 `api/gen/`。

## 已通的主流程

1. 注册登录后，经网关用 HTTP 访问商品、购物车、库存、订单、支付、优惠券、搜索和站内信。
2. 下单带 `Idempotency-Key`。可选 `user_coupon_id`：先经 gRPC 锁券，再经 gRPC 预留库存，应付金额为商品额减去减免。
3. 取消待支付订单会经 gRPC 归还库存并解锁券。
4. mock 确认支付后，支付服务经 gRPC 调用订单 `MarkPaid`。订单经 gRPC 确认库存并核销券。
5. 下单、取消、支付成功各经 gRPC 写一条站内信。通知失败只记日志，不让下单失败。
6. 创建商品或 SKU 后，经 gRPC 更新搜索索引。搜索服务没启动时，商品写入仍然成功。

浏览器不直接访问这些 gRPC 端口。HTTP 上的 lock / unlock / confirm、通知写入、搜索写入、库存预留释放确认、`POST /orders/:id/paid` 仍在，给手工调试用。

## 还没做

user、cart、gateway 的 gRPC（当前没有服务间调用方）、NATS、OpenSearch、超时关单、从购物车下单、真实支付渠道。
