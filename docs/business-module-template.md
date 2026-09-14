# 业务模块开发模板

新增业务能力时，后端、前端、数据库、接口契约和测试必须作为一个纵向切片交付。不要先创建全局 `handlers`、`services`、`repositories` 或 `utils` 目录。

## 后端目录

```text
api/internal/modules/<module>/
├── application/          用例、端口和应用服务
├── domain/               实体、值对象和业务不变量
├── persistence/          数据库仓储实现
├── transport/http/       Gin Handler 和 DTO 映射
├── application/*_test.go 用例测试
└── module.go             构造、权限、资源和路由的公开边界
```

模块必须通过接口依赖仓储，不得把数据库模型直接作为 API 响应。跨模块访问通过公开用例、端口或事件完成。

## 接入步骤

1. 先在 `docs/contracts/openapi.yaml` 增加 `/v1/` 接口、参数、响应、权限和错误码。
2. 添加 `api/migrations/<序号>_<module>.sql`，迁移只能前向执行，不修改已经应用的文件。
3. 实现领域模型、用例、仓储和 Handler，并为成功、未找到、校验失败、未授权和依赖失败写测试。
4. 在模块边界中提供 `RegisterRoutes`、`RegisterPermissions` 和必要的 `RegisterResources`。
5. 在 `api/cmd/server/main.go` 的组合根中只负责构造依赖和注册模块，不放业务规则。
6. 前台业务放在 `web/src/business/<module>/`，并只加入 `web/src/app/public-business-modules.js`；管理后台业务放在 `web/admin/src/business/<module>/`，并只加入 `web/admin/src/app/admin-business-modules.js`。两者不得互相导入。页面必须处理 loading、empty、error、retry、unauthorized 和 mutation-in-progress 状态。
7. 更新对应的模块文档、OpenAPI 字段语义和关键路径 E2E。

## 验收清单

- API 使用 `/v1/`、camelCase JSON、`data`/`meta` 响应封装和稳定错误码。
- 列表查询限制大小，排序和筛选字段白名单化，时间使用 UTC RFC3339。
- 所有需要登录的路由明确声明权限；写操作记录审计事件。
- 多写操作使用事务；查询使用 `context.Context` 和明确超时。
- 前端处理 loading、empty、error、retry、unauthorized 和 mutation-in-progress 状态。
- Go 测试、前端测试、OpenAPI 校验和关键路径 E2E 均纳入交付验证。
