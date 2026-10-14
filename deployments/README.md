# deployments

Docker Compose templates only. They intentionally have no runnable service definitions until M0 is implemented.

- `docker-compose.yml`: application and infrastructure composition
- `docker-compose.obs.yml`: Prometheus, Grafana, Loki, Tempo and OTel Collector
- `.env.example`: non-secret environment-variable template
- `mysql/init/`: future database bootstrap SQL
- `prometheus/`: future scrape and alert configuration

Read [docs/10-deployment.md](../docs/10-deployment.md) before adding containers.
