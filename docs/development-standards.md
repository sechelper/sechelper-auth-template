# 开发规范

## 默认视觉设计

除非需求、产品设计或用户明确指定其他视觉风格，所有新建或改造的 React 页面、组件和交互状态默认采用仓库内的 Sneat Bootstrap FREE 设计参考：

[`docs/ai-design-reference/sneat-bootstrap-free/`](ai-design-reference/sneat-bootstrap-free/)

这里的“采用”指遵循该参考中的信息层级、布局模式、色彩令牌、间距、圆角、阴影、组件状态和后台信息密度；不要求原样复制 Bootstrap 或引入完整 Sneat/Bootstrap 运行时。具体实施前先阅读：

- [`README.md`](ai-design-reference/sneat-bootstrap-free/README.md)：适用范围和文件索引
- [`design-tokens.css`](ai-design-reference/sneat-bootstrap-free/design-tokens.css)：颜色、字体、尺寸和基础组件令牌
- [`layout-patterns.md`](ai-design-reference/sneat-bootstrap-free/layout-patterns.md)：后台壳层、菜单、导航、容器和栅格
- [`component-catalog.md`](ai-design-reference/sneat-bootstrap-free/component-catalog.md)：组件结构、状态和使用场景
- [`forms-alerts.md`](ai-design-reference/sneat-bootstrap-free/forms-alerts.md)：表单、校验、告警和反馈

### 优先级和适用范围

视觉决策按以下优先级处理：

1. 用户在当前需求中明确指定的样式、品牌或设计稿；
2. 项目已有且仍在使用的页面主题约定；
3. 本规范规定的 Sneat 参考风格。

本默认约定适用于 `web/`、`web/admin/` 中的新页面和新组件，也适用于登录页、管理后台、数据列表、表单、空状态、加载态、错误态和成功反馈。已有页面不会因新增本规范而自动整体改版；只有在该页面被改造时，才按本规范逐步统一受影响的区域。

### 实现要求

- 以 `docs/ai-design-reference/sneat-bootstrap-free/design-tokens.css` 为视觉令牌来源，在目标前端的统一样式入口建立或复用对应 token；组件内不得随意创造近似颜色、间距或阴影。
- 默认使用浅灰蓝页面画布（`#f5f5f9`）、白色内容卡片、`#696cff` 紫色主操作色，以及低强度阴影和紧凑圆角；危险、成功、警告和信息状态使用参考目录中对应的语义色。
- 后台页面优先采用固定侧边菜单、顶部导航、内容容器和卡片化信息区；常规页面遵循“页面标题/面包屑 → 关键指标或操作 → 主内容卡片 → 辅助内容 → 分页或底部动作”的层次。
- 一个首屏保持一个视觉主焦点。避免并列堆叠多个紫色实心主按钮、过度渐变、强烈发光、彩虹配色和大面积玻璃拟态。
- 组件必须覆盖可用的加载、空、错误、未授权、重试和操作进行中状态；状态不能只依赖颜色，还要提供文字、图标、形状或可访问标签。
- 保留 `SECHELPER COMMUNITY` 品牌锁定、原始 logo 比例和项目既有认证边界。Sneat 仅定义默认视觉语言，不得改变认证、权限、接口或数据规则。
- 默认桌面 Web 布局；只有需求明确包含移动端、手机或 H5 时，才增加移动端断点和专用交互。
- 遵守 WCAG AA，对键盘操作提供清晰的 `2–3px` focus outline，避免缩放或长文本造成重叠；颜色对比度不足时以可读性为准调整。
- 参考目录只存设计资料，不放业务接口、权限判断、用户数据或项目特有文案。业务逻辑仍由对应前端模块和 Go API 负责。

### 交付检查

实现涉及视觉变更时，应在交付说明中指出使用了 Sneat 参考风格，并简要说明主色、页面画布、卡片、状态色和焦点样式的使用位置。若采用其他风格，应说明覆盖了哪条默认约定及原因，并确保该例外只影响明确的页面或功能范围。
