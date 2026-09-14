# 表单组件与交互状态

## 基础规则

表单沿用 Bootstrap 5：`.form-label`、`.form-control`、`.form-select`、`.input-group`、`.form-check`、`.form-switch`、`.btn-check`。默认控件字体继承 Public Sans 15px，边框为 `#E4E6E8`，圆角默认 6px，内边距约 8px 14px；placeholder 使用辅助灰。

| 类型 | 规范 |
|---|---|
| Input / Select / Textarea | 白底、1px 浅灰边框；focus 主色边框与 ring；textarea 垂直可调整（Bootstrap 默认） |
| Checkbox / Radio | 使用 `.form-check-input`；checked 主色；label 与控件垂直居中 |
| Switch | `.form-switch`；轨道灰/主色，圆形 thumb；disabled 降低透明度 |
| Date picker | 原生/第三方实现待确认；推荐复用 input 尺寸、图标槽位和 focus 状态，不在资料中虚构日历面板 |
| Validation | `.is-valid` 使用 `#71DD37`；`.is-invalid` 使用 `#FF3E1D`；反馈文字同状态色，小号文字；错误边框优先级高于默认边框 |
| Disabled | `:disabled` 降低 opacity、禁止 pointer；背景偏灰，文字辅助灰 |
| Input group | `.input-group-text` 与控件共用边框和圆角，仅外侧保留圆角；图标/前后缀色为辅助灰 |

验证页面链接是 Pro：`form-validation.html`，Free 版本能确认基础变量但不能将 Pro 的字段组合和验证流程当作已采集事实。
