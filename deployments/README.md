# deployments

Docker Compose 模板目录。在实现 M0 之前，故意不包含可运行的服务定义。

- `docker-compose.yml`：应用与基础设施编排
- `docker-compose.obs.yml`：Prometheus、Grafana、Loki、Tempo、OTel Collector
- `.env.example`：非敏感环境变量模板
- `mysql/init/`：未来的数据库初始化 SQL
- `prometheus/`：未来的抓取与告警配置

添加容器前请先阅读 [docs/10-deployment.md](../docs/10-deployment.md)。