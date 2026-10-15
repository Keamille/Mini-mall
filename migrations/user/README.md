# user 迁移

- 所属数据库：`minimall_user`
- `000001_initial.up.sql` / `000001_initial.down.sql`：users、user_addresses、refresh_tokens

本地快速应用（Compose 已启动时）：

```powershell
Get-Content migrations/user/000001_initial.up.sql -Raw |
  docker compose -f deployments/docker-compose.yml exec -T mysql `
    mysql -uminimall -p123456 minimall_user
```
