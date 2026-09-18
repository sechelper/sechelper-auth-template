# 开发环境与质量门禁

## 统一构建与启动

项目使用统一的配置载入入口和程序启动入口，环境差异由显式的 `ENV` 参数选择，不为测试或示例代码增加独立的启动入口。构建与运行命令如下：

```bash
make build ENV=development
make run ENV=development ACTION=serve
make run ENV=development ACTION=migrate
```

`ENV` 可取 `development`、`test` 或 `production`。`COMPONENT=all|api|web` 可缩小构建范围；`ACTION=serve|migrate` 选择统一入口中的程序行为。`make dev` 和 `make db-migrate` 是面向开发者的快捷别名，最终仍委托统一的运行流程。

开发环境默认使用专用 Compose project，地址与服务组成以 [deploy/compose.dev.yaml](../deploy/compose.dev.yaml) 及配置样例为准。测试环境业务服务按远程物理机规范使用主机 Go `1.26.8`、Node.js `24.21.0` 和 npm `11.19.1` 原生构建，PostgreSQL/Redis 等基础设施由 [deploy/compose.test.yaml](../deploy/compose.test.yaml) 运行；生产环境业务服务由 `deploy/compose.prod.yaml` 运行。环境名通过统一 Make 入口传入，不再维护项目环境文件。不要在生产配置中启用示例模块或测试迁移。

API 与迁移程序均通过 `--config` 指定配置文件。配置由统一配置包载入，并校验配置环境与构建环境一致；不得在业务代码中自行读取环境变量、创建另一套配置加载方式或绕过启动入口。

## 迁移隔离

业务迁移按模块目录组织。每个模块可通过 `MODULE_KIND` 声明 `production` 或 `example`；未声明时按生产模块处理。生产构建只打包生产迁移，测试/开发构建可包含示例迁移。迁移程序在生产环境拒绝示例迁移，即使文件意外存在也不会执行。框架迁移写入 `framework` schema，配置中心迁移写入 `configuration` schema，`business/<module>/` 迁移写入 `business_<module>` schema。迁移账本为 `framework.schema_migrations`，版本键为 `(scope, version)`；业务 SQL 不得依赖 `public` schema 或未限定的跨模块表。

## 常用验证

```bash
make architecture-check
make test
make openapi-lint
make build ENV=test
```

运行测试和构建前遵守 [测试环境说明](testing-environment.md)。不得在仓库留下二进制、测试缓存、覆盖率文件、运行日志或其他构建产物。生产构建仅用于正式发布流程；本地开发和测试不得连接或操作生产服务及数据。

## 新业务模块

新模块遵循[业务模块开发模板](business-module-template.md)，通过框架注册器接入权限、资源、审计事件和受保护路由。OpenAPI、迁移、实现、测试和文档均需归属同一个业务模块。CI 执行格式、测试、前后台构建、合同校验与架构检查；不得通过修改检查脚本或 CI 策略来规避失败。
