# Sneat Bootstrap Free 视觉设计参考

本目录是对 [Sneat Bootstrap HTML Admin Template Free](https://demos.themeselection.com/sneat-bootstrap-html-admin-template-free/html/index.html) 的可复用视觉规范提取，按模板左侧导航顺序组织。资料采集日期：2026-09-13；来源页面以公开 Free Demo 为准。

## 使用方式

- 先阅读 `design-tokens.md`、`layout.md`、`navigation.md`，再按需要组合 `components.md` 和 `forms.md`。
- `pages/00-dashboard-analytics.md` 至 `pages/44-tables-basic.md` 与左侧菜单的 Free 可访问页面保持顺序；Pro 项目保留在 `navigation.md` 的外链表中，不将 Pro 页面内容冒充为 Free 页面。
- `assets/core.css`、`assets/demo.css`、`assets/iconify-icons.css` 是当次采集的样式快照；图片文件是首页可见资源。若本地复刻，优先使用 CSS 变量和类名，而不是硬编码截图像素。
- “推断”表示由 CSS/DOM/截图归纳；“待确认”表示公开页面无法直接验证或需要在特定浏览器/交互状态下补采。

## 证据与范围

| 项目 | 证据 |
|---|---|
| 基准页面 | `html/index.html` |
| 样式入口 | `assets/vendor/css/core.css`、`assets/css/demo.css`、`assets/vendor/fonts/iconify-icons.css`、Perfect Scrollbar、ApexCharts |
| 技术基线 | Bootstrap 5.3.3；页面 DOM 使用 `.layout-wrapper`、`.layout-menu`、`.layout-page`、`.layout-navbar`、`.content-wrapper`、`.card` |
| 字体 | Public Sans，系统字体回退链见 `design-tokens.md` |
| 视觉对照 | `screenshots/dashboard-desktop.png` |

## 复刻原则

1. 默认采用浅色主题；深色变量在 `core.css` 中存在，但 Free Demo 当前页面不是深色展示。
2. 采用 4px 基础间距节奏，Bootstrap gutter 默认为 1.625rem（26px）。
3. 颜色、圆角、阴影和断点从源码优先；页面实际尺寸以浏览器 viewport 和具体容器为准。
4. Pro 菜单项仅作为导航占位，必须显示 `PRO` 徽标并使用外链；不要把 Pro 功能实现归入 Free 页面。
