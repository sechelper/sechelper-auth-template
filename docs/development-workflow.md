# 开发环境与质量门禁

## 本地启动

项目使用专用 Compose project 启动 PostgreSQL、Redis、开发用 Mock Identity Provider、数据库迁移、API 和前端容器。API 依赖迁移服务成功完成后才启动：

```bash
make dev
```

默认地址：

- 前台：http://localhost:8088
- API：http://localhost:8080
- 健康检查：http://localhost:8080/healthz
- 就绪检查：http://localhost:8080/readyz

开发环境的身份平台地址是 Compose 内部的 `mock-idp:9000`，浏览器登录仍从 `localhost:8080` 回调到 API。Mock Identity Provider 只存在于开发 Compose，不得用于生产部署。

开发 Compose 使用项目专用的容器、网络和命名卷。停止服务但保留数据使用 `make dev-down`；清理数据时必须显式确认目标是本项目的开发 Compose project。

Compose 开发栈使用 [deploy/config.dev.yaml](../deploy/config.dev.yaml)。手动运行 API 时，复制 `config.example.yaml` 为项目根目录的 `config.yaml`，只在本地填写非生产配置；秘密通过环境变量传入，不要提交 `config.yaml`。

## 常用验证

```bash
make test
make build
make openapi-lint
make docker-build
```

Go 的模块测试与生产构建不应把二进制、测试缓存、覆盖率文件或运行日志写回仓库。生产环境必须显式传入 `APP_CONFIG_FILE`，并在发布步骤单独执行数据库迁移。

## 新业务模块

新模块必须遵循 [业务模块开发模板](business-module-template.md)，先更新 OpenAPI 和迁移，再实现后端与前端，最后补测试和文档。CI 会检查 Go 格式、Go 测试、前后台构建、OpenAPI 和生产镜像构建。
