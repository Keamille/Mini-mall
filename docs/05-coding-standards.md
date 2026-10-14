# 05 编码规范

## 1. 通用约定

- 格式化：`gofmt` / `goimports`；CI 检查 `golangci-lint`
- 命名：包名小写单词；导出名驼峰；文件名小写+下划线按需
- Context：入口传入 `context.Context`，禁止在 struct 中长期存可取消 ctx
- 错误：返回时包装业务语义，`fmt.Errorf("...: %w", err)`；日志处记录一次即可
- 禁止在库代码中 `panic`（除明确不可恢复的启动失败）

## 2. 分层与依赖注入

- `domain` 不依赖基础设施
- `app` 依赖 `port` 接口；具体实现在 `adapter`
- 启动时在 `server` / `main` 组装，便于测试替换 Mock

## 3. 统一错误码

错误体见 [07-api-conventions.md](07-api-conventions.md)。  
**完整分段与细码总表**见 [15-error-codes.md](15-error-codes.md)（单一事实源）；此处仅保留分段提醒：

| 范围 | 含义 |
|------|------|
| 0 | 成功 |
| 10000–10999 | 通用 |
| 11000–11999 | user |
| 12000–12999 | product |
| 13000–13999 | inventory |
| 14000–14999 | cart |
| 15000–15999 | order |
| 16000–16999 | payment |
| 17000–17999 | coupon |
| 18000–18999 | search |
| 19000–19999 | notification |

新增业务码时同步更新 [15-error-codes.md](15-error-codes.md) 与对应服务文档。

## 4. 日志

- 使用结构化日志（`slog` 或 zap）
- 必带字段：`trace_id`、`request_id`、`user_id`（若有）、`service`
- 级别：业务可预期失败用 `Warn`；未知用 `Error`
- 禁止记录密码、完整银行卡、Token、支付密钥

## 5. 时间与金额

- 对外时间：RFC3339，UTC 存储
- 金额：整数**分**（`int64`），禁止 `float`
- 货币：一期固定 `CNY` 字段预留

## 6. ID 策略

- 对外业务 ID：**统一 ULID**（可排序、字符串友好）
- DB 主键：`CHAR(26)` 存 ULID
- 订单号、支付号：前缀 `O`/`P` + ULID（或时间有序段），便于排查

## 7. 测试

| 类型 | 要求 |
|------|------|
| 单测 | domain 与纯函数必测；app 用例用 port mock |
| 集成测 | `testing` + testcontainers 或 Compose 测试 profile |
| 契约 | proto breaking change 在 CI 检测（buf breaking） |

覆盖率不设教条阈值；交易路径（库存预留/确认/释放、支付回调幂等）必须有测试。

## 8. Git 与提交

- 分支：`main` + `feature/*` / `fix/*`
- Commit 信息：简明说明动机（中英均可，团队统一即可）
- 禁止提交 `.env`、密钥、大型二进制

## 9. 相关文档

- API 约定：[07-api-conventions.md](07-api-conventions.md)
- 错误码总表：[15-error-codes.md](15-error-codes.md)
- 安全：[12-security.md](12-security.md)

---

导航：[上一篇 04 目录结构](04-project-structure.md) | [下一篇 服务索引](06-services/README.md) | [返回索引](README.md)
