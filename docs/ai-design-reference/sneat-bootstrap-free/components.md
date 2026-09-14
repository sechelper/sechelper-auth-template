# 通用组件规范

| 组件 | 结构/类名 | 视觉与状态 |
|---|---|---|
| Button | `.btn`, `.btn-primary`, `.btn-outline-primary`, `.btn-icon` | 6px 圆角；主色实心白字；outline 主色边框；disabled 降低不透明度；focus 使用 ring |
| Card | `.card`, `.card-header`, `.card-body`, `.card-footer` | 白底、6px、`0 3px 8px rgba(34,48,62,.10)`；标题 18px/500，内边距按 Bootstrap card token |
| Badge | `.badge`, `.rounded-pill`, `.bg-*` | 小号约 13px、500；胶囊；状态色实心或浅色变体 |
| Alert | `.alert.alert-*` | 状态色浅背景、状态色文字/边框；可关闭 |
| Table | `.table`, `.table-striped`, `.table-hover` | 1px 浅灰分隔；表头约 13px/500，内容 15px；hover 为浅色背景 |
| Pagination | `.pagination`, `.page-item`, `.page-link` | 6px/圆角；active 主色；disabled 辅助灰 |
| Dropdown | `.dropdown-menu`, `.dropdown-item` | 白底、圆角、阴影；item hover 主色浅底；三点操作常用 `btn-icon` |
| Modal | `.modal`, `.modal-dialog`, `.modal-content` | 白底、8px 左右；遮罩深色半透明；header/body/footer 分区 |
| Breadcrumb | `.breadcrumb` | 13–15px；当前项深色，分隔符浅灰 |
| Progress | `.progress`, `.progress-bar` | 8px 左右高度、胶囊端点；主色/状态色填充 |
| Spinner | `.spinner-border`, `.spinner-grow` | 默认主色；可用 `.text-*` 变体 |
| Tabs | `.nav-tabs`, `.nav-pills`, `.nav-link` | active 主色；pills 6px 圆角，tab 内容白底卡片化（依页面） |
| Toast | `.toast` | 白底、阴影、右上/右下浮层；状态 icon/色条依变体 |
| Tooltip/Popover | Bootstrap tooltip/popover | 深色 tooltip；popover 白底阴影和标题/正文分层 |

## 数据可视化

首页使用 ApexCharts（`apex-charts.css`、`dashboards-analytics.js`）；图表颜色应从主色和状态色派生，背景保持透明/卡片白底。图表 tooltip、坐标轴、网格的最终值属于运行时 ApexCharts 配置，需以对应页面 JS 为准（待确认）。
