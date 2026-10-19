.PHONY: help proto migrate-up test docker-up docker-down

help:
	@echo "当前仅为项目骨架，请先实现脚本后再使用下列目标。"

proto:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/gen-proto.ps1

migrate-up:
	@echo "待实现：执行数据库迁移"

test:
	@echo "待实现：运行测试"

docker-up:
	@echo "待实现：启动 Docker Compose 服务"

docker-down:
	@echo "待实现：停止 Docker Compose 服务"