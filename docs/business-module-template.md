# 业务模块开发模板

授权注册和框架维护的权威边界见[授权框架维护](authorization-framework-maintenance.md)。本文档只保留业务模块纵向切片的实施步骤；当改动注册协议、授权语义或宿主组合根时，先按该文档升级为框架维护任务。

新增业务能力时，后端、前端、数据库、接口契约和测试必须作为一个纵向切片交付。不要先创建全局 `handlers`、`services`、`repositories` 或 `utils` 目录。

同一业务能力必须在后端、业务前台和管理前台使用完全相同的 `<module>` 目录名：`api/internal/business/<module>/`、`web/src/business/<module>/`、`web/admin/src/business/<module>/`。不需要某一端时可以不创建空目录，但不得使用别名拆散同一个业务模块。

## 后端目录

```text
api/internal/business/<module>/
├── application/          用例、端口和应用服务
├── domain/               实体、值对象和业务不变量
├── persistence/          数据库仓储实现
├── transport/http/       Gin Handler 和 DTO 映射
├── application/*_test.go 用例测试
└── module.go             构造、权限、资源和路由的公开边界
```

模块必须通过接口依赖仓储，不得把数据库模型直接作为 API 响应。跨模块访问通过公开用例、端口或事件完成。

## 接入步骤

1. 先在 `docs/contracts/business/<module>/` 定义 `/v1/` 接口、参数、响应、权限和错误码；需要合入主 OpenAPI 时提交框架扩展申请。
2. 添加 `api/migrations/business/<module>/<序号>_<module>.sql`，迁移只能前向执行，不修改已经应用的文件。
3. 实现领域模型、用例、仓储和 Handler，并为成功、未找到、校验失败、未授权和依赖失败写测试。
4. 在模块边界实现框架的 `BusinessModule` 注册契约：提供 `Name`、`RegisterRoutes`、`RegisterPermissions`、`RegisterResources`、`RegisterAuditEvents`；无资源或审计事件时也显式实现空注册。路由通过宿主提供的授权接口声明权限。
5. 业务模块只提供公开构造和注册声明；需要修改 `api/cmd/server/main.go` 或后端框架注册器时，按[授权框架维护](authorization-framework-maintenance.md)提交框架维护申请，不得在业务任务中直接修改组合根。
6. 真实前台业务放在 `web/src/business/<module>/` 并由 `public-module.js` 自描述路由；启用真实业务模块中最多一个可以声明 `/`，正式构建必须恰有一个真实业务首页。开发/测试演示使用 `example-module.js` 声明，只允许在 Vite development 或显式 `test` 模式动态载入，正式构建不得发现或打包这些模块；演示模块不能替代真实首页。管理后台业务放在 `web/admin/src/business/<module>/` 并由 `admin-module.js` 自描述路由、导航、顺序和权限。框架会自动发现声明，不得为接入模块修改主文件、全局路由或注册器；只有改变声明协议或发现机制时才提交框架扩展申请。前后台不得互相导入，页面必须处理 loading、empty、error、retry、unauthorized 和 mutation-in-progress 状态。
7. 更新对应的模块文档、OpenAPI 字段语义和关键路径 E2E。

## 验收清单

- API 使用 `/v1/`、camelCase JSON、`data`/`meta` 响应封装和稳定错误码。
- 列表查询限制大小，排序和筛选字段白名单化，时间使用 UTC RFC3339。
- 所有需要登录的路由明确声明权限；写操作记录审计事件。
- 多写操作使用事务；查询使用 `context.Context` 和明确超时。
- 前端处理 loading、empty、error、retry、unauthorized 和 mutation-in-progress 状态。
- 正式前台构建必须注册且只注册一个真实业务根路由 `/`；演示模块不计入该路由检查，缺少真实业务首页时构建失败，正式运行不得显示模板或演示页面。
- 管理后台的每条业务路由和导航声明相同的服务端权限码，导航路径必须对应本模块路由；模块名、路由和导航路径不得重复。
- 管理后台导航支持递归 `children`。同级分组默认仅作为标题展示、不折叠；需要分组自身折叠时显式设置 `sectionCollapsible: true`。业务节点可继续嵌套，当前页面所在节点由框架自动展开。
- Go 测试、前端测试、OpenAPI 校验和关键路径 E2E 均纳入交付验证。
