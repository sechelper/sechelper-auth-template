# 左侧导航系统

以下顺序严格按首页 DOM 中的左侧导航读取。一级分组标题是视觉分隔，不是链接。

1. **Dashboards**（徽标 `5`，图标 `bx-home-smile`）：Analytics；CRM / eCommerce / Logistics / Academy 均为 Pro 外链。
2. **Layouts**（`bx-layout`）：Without menu；Without navbar；Fluid；Container；Blank。
3. **Front Pages**（`bx-store`，`PRO`）：Landing；Pricing；Payment；Checkout；Help Center（Pro 外链）。
4. **Apps & Pages**：Email（`bx-envelope`，Pro）；Chat（`bx-chat`，Pro）；Calendar（`bx-calendar`，Pro）；Kanban（`bx-grid`，Pro）。
5. **Account Settings**（`bx-dock-top`）：Account；Notifications；Connections。
6. **Authentications**（`bx-lock-open-alt`）：Login；Register；Forgot Password。
7. **Misc**（`bx-cube-alt`）：Error；Under Maintenance。
8. **Components** 分组：Cards（`bx-collection`）；User interface（`bx-box`）：Accordion；Alerts；Badges；Buttons；Carousel；Collapse；Dropdowns；Footer；List groups；Modals；Navbar；Offcanvas；Pagination & Breadcrumbs；Progress；Spinners；Tabs & Pills；Toasts；Tooltips & Popovers；Typography；Extended UI（`bx-copy`）：Perfect Scrollbar；Text Divider；Boxicons（`bx-crown`）。
9. **Forms & Tables**：Form Elements（`bx-detail`）：Basic Inputs；Input groups；Form Layouts（`bx-detail`）：Vertical Form；Horizontal Form；Form Validation（`bx-list-check`，Pro）；Tables（`bx-table`）；Datatables（`bx-grid`，Pro）。
10. **Misc**（底部）：Support（`bx-support`）；Documentation（`bx-file`）。

## 状态与层级

- 一级链接类名：`.menu-link`; 可展开项附 `.menu-toggle`，展开态由 `.open`/父级菜单状态和 `aria-expanded`（以具体脚本状态为准）控制。
- 当前 Analytics 使用 `.active`：淡紫蓝背景（约 `rgba(105,108,255,.16)` 的视觉效果）、主色文字和图标，圆角约 6px。
- 悬停通常采用主色浅背景/主色文字；未选中菜单是深灰文字，子项通过缩进和小圆点区分。
- Pro 徽标是淡紫背景、主色文字、胶囊圆角、小号大写字样；Dashboards 数字徽标为红色胶囊白字。
- 侧栏品牌为 “Sneat” 文字 Logo；收起状态使用 `.app-brand-img-collapsed`，公开首页未提供独立文本 SVG，待确认具体图形资产。
- 图标体系使用 Boxicons/Iconify CSS 快照，常见类名为 `bx bx-*` 和 `menu-icon tf-icons`。
