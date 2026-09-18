# `orders` 测试模块清理快捷文档

## 目的

新业务接入前，用本文档快速识别并清除仓库中的 `orders` 示例/测试业务模块，避免模板业务、测试接口、测试迁移和示例名称进入正式项目。

本文档只适用于当前仓库中已标记为 `example` 的 `orders` 模块。若目标内容已被改造成正式业务，必须先停止清理并由用户明确确认处理范围。

## 清理前确认

开始删除前必须确认：

1. 当前接入业务不依赖 `orders` 的 API、权限、数据表、管理页面或示例合同。
2. `orders` 相关迁移仍属于示例/测试用途，未作为正式业务数据结构使用。
3. 工作区已有修改已记录，清理过程中不得覆盖、还原或整理无关变更。
4. 不得删除框架自身的通用测试；只清理明确属于 `orders` 示例模块的内容。

## 清理范围

按以下范围检查并清除，实际路径以全量检索结果为准：

- 后端模块：`api/internal/business/orders/**`
- 示例迁移：`api/migrations/business/orders/**`
- 管理端业务页面：`web/admin/src/business/orders/**`
- 后台开发参考模块：`web/admin/src/business/component-reference/**`
- 业务对接文档：`docs/business/orders/integration.md`、`docs/orders.md` 等示例说明
- 业务接口合同：`docs/contracts/business/orders/**`
- 清理指南：保留本文档，作为后续新业务接入的快捷清理依据
- 示例运行时装配、示例构建入口和相关注册引用
- `orders`、`order:read`、`example-order`、`/v1/orders`、`/admin/orders` 等仅服务于该示例模块的名称、权限、路由和配置

## 快速执行顺序

1. 全量检索 `orders`、`order:read`、`example-order`、`/v1/orders`、`/admin/orders`、`component-reference`、`/admin/component-reference`、`开发参考` 及明显的模板项目名。
2. 区分正式框架代码、通用测试和 `orders` 示例模块，建立待清理文件清单。
3. 删除 `orders` 业务目录、示例迁移、业务合同、对接文档、后台“开发参考”模块及其仅用于示例装配的引用；保留本文档，不得将清理指南作为待删除的业务文档。
4. 清理示例模块和“开发参考”模块对应的权限、资源、路由、菜单、构建和迁移引用；不得修改通用框架能力来绕过检查。
5. 再次全量检索，确认没有 `orders` 示例残留、`order:read` 权限残留、示例路由残留、“开发参考”模块残留或旧模板名残留。
6. 检查完整 Git 差异，确认只删除目标示例模块和必要引用，没有影响正式业务或框架基线。

## 验收标准

- 不存在 `api/internal/business/orders/`、`api/migrations/business/orders/` 及对应前台、后台业务目录。
- 不存在 `orders` 示例 API、管理端路由、菜单、权限、资源定义和数据库对象。
- 不存在 `web/admin/src/business/component-reference/`、`/admin/component-reference` 路由、`开发参考` 菜单及组件参考页面。
- 不存在仅用于 `orders` 示例的测试数据、测试配置、测试迁移、示例合同和对接文档；本文档作为清理依据保留。
- 正式构建、正式迁移和部署配置不再引用该示例模块。
- 完成格式检查、相关测试、合同检查和 `make architecture-check`。
- 清理后应同步更新业务接入文档，确保新业务使用自己的正式名称，不得继续沿用 `orders` 或其他模板名称。

## 风险提示

不得使用全仓库通配删除，也不得根据字符串匹配直接删除所有包含 `order` 的文件。`order` 可能出现在通用排序字段、通用测试名称或正式业务语义中，必须结合文件所属模块和用途逐项确认。
