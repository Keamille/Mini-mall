# 10 Docker 部署

一期以 **Docker Compose** 为开发/演示主路径。本文描述目标拓扑与约定；实际 `deployments/docker-compose.yml` 在代码阶段落地。

## 1. 网络与卷

| 资源 | 名称 | 说明 |
|------|------|------|
| 网络 | `minimall_net` | bridge，服务间 DNS 用 Compose 服务名 |
| 卷 | `mysql_data` | MySQL 数据 |
| 卷 | `redis_data` | Redis（可选持久化） |
| 卷 | `nats_data` | JetStream 存储 |
| 卷 | `opensearch_data` | 搜索数据 |

不建议把业务容器做成有状态；状态在中间件卷中。

## 2. Compose 服务清单

### 2.1 基础设施（profile: `infra`）

| 服务名 | 镜像（建议） | 端口映射 | 说明 |
|--------|--------------|----------|------|
| `mysql` | `mysql:8.4` | `3306:3306` | 多 database 初始化脚本 |
| `redis` | `redis:7-alpine` | `6379:6379` | |
| `nats` | `nats:2.10` | `4222:4222`, `8222:8222` | `-js -m 8222` |
| `opensearch` | `opensearchproject/opensearch:2` | `9200:9200` | 单节点；关安全插件便于本地 |
| `opensearch-dashboards` | 可选 | `5601:5601` | profile `search-ui` |

### 2.2 可观测（profile: `obs`）

| 服务名 | 说明 |
|--------|------|
| `prometheus` | 抓取各服务 `/metrics` |
| `grafana` | 仪表盘 |
| `loki` + `promtail` | 日志 |
| `tempo` | Trace 后端（一期默认） |
| `otel-collector` | 接收 OTLP，转发 Prometheus/Tempo/Loki |

### 2.3 业务服务（profile: `app` 或默认）

端口、依赖的**完整矩阵**见 [14-port-matrix.md](14-port-matrix.md)。摘要：

| 服务名 | 依赖 | 端口 |
|--------|------|------|
| `gateway` | redis + 可路由 svc | `8080` |
| `user-svc` | mysql | gRPC `9001` |
| `product-svc` | mysql, nats | `9002` |
| `cart-svc` | mysql, product | `9003` |
| `order-svc` | mysql, nats, 多 svc | `9004` |
| `payment-svc` | mysql, nats | `9005`, HTTP `8085` |
| `coupon-svc` | mysql, nats | `9006` |
| `search-svc` | opensearch, nats | `9007` |
| `inventory-svc` | mysql, nats | `9008`（不对公网） |
| `notification-svc` | mysql, nats, user | `9009` |

业务镜像：多阶段构建，仓库内 `Dockerfile` 按 `CMD` 参数区分或每服务一个 Dockerfile。

## 3. 分阶段启动

```bash
# 1) 中间件
docker compose --profile infra up -d

# 2) 迁移（代码阶段提供）
# make migrate-up

# 3) 核心交易（可先不含 search/notification）
docker compose --profile infra --profile app up -d gateway user-svc product-svc inventory-svc cart-svc order-svc payment-svc coupon-svc

# 4) 搜索与通知
docker compose up -d search-svc notification-svc opensearch

# 5) 可观测
docker compose --profile obs up -d
```

健康检查：

- MySQL：`mysqladmin ping`
- Redis：`redis-cli ping`
- NATS：监控口 `/healthz`
- 业务：HTTP `/healthz` 或 gRPC Health

`depends_on` 配合 `condition: service_healthy`。

## 4. MySQL 初始化

挂载 `deployments/mysql/init`：

```sql
CREATE DATABASE IF NOT EXISTS minimall_user CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_product CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_inventory CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_cart CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_order CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_payment CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_coupon CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS minimall_notification CHARACTER SET utf8mb4;
```

应用账号最小权限：每库独立用户（演示可用同一 `minimall` 用户）。

## 5. 环境变量模板（`.env.example` 说明）

```bash
# 全局
MINIMALL_ENV=dev
MYSQL_ROOT_PASSWORD=changeme
MINIMALL_MYSQL_USER=minimall
MINIMALL_MYSQL_PASSWORD=changeme
MINIMALL_MYSQL_HOST=mysql
MINIMALL_MYSQL_PORT=3306

# JWT（演示可用 HS256 共享密钥；生产改 RS256）
MINIMALL_JWT_HS_SECRET=dev-only-change-me

# NATS
MINIMALL_NATS_URL=nats://nats:4222

# Redis
MINIMALL_REDIS_ADDR=redis:6379

# OpenSearch
MINIMALL_SEARCH_OPENSEARCH_URL=http://opensearch:9200

# Payment
MINIMALL_PAYMENT_MOCK=true

# Gateway
MINIMALL_GATEWAY_HTTP_ADDR=:8080
```

各服务 DSN 示例：

```text
MINIMALL_USER_MYSQL_DSN=minimall:changeme@tcp(mysql:3306)/minimall_user?parseTime=true&loc=UTC
```

**禁止**将真实生产密钥提交到 Git。

## 6. 资源建议（本地）

| 组件 | CPU | 内存 |
|------|-----|------|
| OpenSearch | 1 | ≥ 1GB（设 `OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m`） |
| MySQL | 1 | 512MB–1GB |
| 单业务 svc | 0.25–0.5 | 64–128MB |

开发机内存紧张时可不用 `obs` profile；搜索仍使用 OpenSearch（可降低堆内存），**不**在一期引入其它搜索引擎。

## 7. 网关入口与回调

- 浏览器/客户端：`http://localhost:8080`
- 支付回调：经 gateway `POST /api/v1/payments/callbacks/{provider}` → payment-svc  
  或临时直暴 `localhost:8085`（仅本地）

## 8. 日志与调试

```bash
docker compose logs -f gateway order-svc payment-svc
docker compose ps
```

## 9. 生产向注意（超出一期 Compose）

- 密钥用 Docker secrets / 外部保管；HTTPS 终止
- MySQL 主从与备份；NATS/OS 集群
- 后续可增加 `deployments/k8s` 附录，不阻塞一期

## 10. 相关文档

- [03-tech-stack.md](03-tech-stack.md)
- [14-port-matrix.md](14-port-matrix.md)
- [11-observability.md](11-observability.md)
- [13-roadmap.md](13-roadmap.md)

---

导航：[上一篇 09 异步事件](09-async-events.md) | [下一篇 11 可观测](11-observability.md) | [返回索引](README.md)
