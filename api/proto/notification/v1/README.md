# notification Proto 契约

在此目录放置 **notification** 域的版本化 protobuf 定义。

- 包名约定：`minimall.notification.v1`
- 生成代码目标目录：`api/gen/`（配置代码生成前可不创建）
- 跨服务 DTO 放在所属域；分页、空消息等公共类型放在 `common/v1`

详见 [API 约定](../../../docs/07-api-conventions.md)。