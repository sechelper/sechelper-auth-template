# 页面布局与响应式

## 桌面结构

```text
body.layout-menu-fixed.layout-compact
└─ .layout-wrapper
   ├─ aside.layout-menu.menu-vertical.bg-menu-theme  (260px)
   └─ .layout-page
      ├─ header.layout-navbar.container-xxl.navbar-detached (白底、6px、顶部 16px、左右 26px)
      ├─ .content-wrapper
      │  └─ main.container-xxl.flex-grow-1.container-p-y
      └─ footer.content-footer.footer.bg-footer-theme
```

桌面端侧边栏固定在左侧，主页面通过 menu width 让出空间；内容卡片为白底、约 6px 圆角、轻阴影。首页截图显示两列统计卡片、宽卡片图表和列表卡片，均由 Bootstrap `.row`/`.col-*` 组织。

## 容器与对齐

- `.container-xxl` 最大宽度：`1440px`（≥1400）；Bootstrap 容器还提供 540/720/960/1140px 档位。
- `.container` 横向 gutter 为 26px，row 子项各承担 13px 内边距。
- navbar detached：页面顶部外边距约 16px，左右约 26px；实际宽度随容器和 viewport 变化。
- 内容区使用 `container-p-y`（源码类，精确值应以当前 CSS 快照为准），卡片间用 `g-*`/`mb-*` 形成规律间距。
- 右侧上下文操作多使用 `btn-icon`、三点菜单、下拉按钮；标题/指标左对齐，数字强调色或深色。

## 响应式

在 `<1200px`：`.layout-menu` 变为 fixed off-canvas，初始 `translate3d(-100%,0,0)`；`.layout-menu-expanded` 时滑入，`.layout-overlay` 显示并使用深色背景、50% 透明度，页面禁止滚动。顶部菜单 toggle 在 xl 以下显示。

在 `≥1200px`：桌面侧栏常驻；`layout-menu-collapsed` 将菜单缩至 `5.25rem`（84px），品牌文字和非悬停图标文字淡出；hover 可临时展开。过渡约 `.3s`。

响应式列遵循 Bootstrap：`col-12` 默认堆叠，在 `md/lg/xl` 逐步切为 `col-md-*`、`col-lg-*`；具体页面以各自源码 class 为准（待确认：未对每一个 Pro 外链页面展开采集）。

## 布局变体

`layouts-without-menu.html`、`layouts-without-navbar.html`、`layouts-fluid.html`、`layouts-container.html`、`layouts-blank.html` 是 Free Demo 中可直接对照的布局变体，见 `pages/01`–`05`。
