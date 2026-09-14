# 设计令牌与代码映射

来源：`assets/core.css`（Bootstrap/Sneat 合并后的公开快照）。以下为浅色主题的主要值。

## 颜色

| 令牌 | 值 | 用途/代码映射 |
|---|---:|---|
| `--bs-primary` | `#696CFF` | 主操作、链接、激活菜单、图表强调；`.btn-primary`、`.text-primary` |
| `--bs-primary-hover` | `#5F61E6` | 链接悬停 |
| `--bs-secondary` | `#8592A3` | 次要操作、辅助状态 |
| `--bs-success` | `#71DD37` | 成功、正向增长、在线 |
| `--bs-info` | `#03C3EC` | 信息、销售/蓝色数据卡 |
| `--bs-warning` | `#FFAB00` | 警告、away |
| `--bs-danger` | `#FF3E1D` | 错误、危险、删除、负向状态 |
| `--bs-dark` | `#2B2C40` | 深色按钮/文字 |
| `--bs-heading-color` | `#384551` | 标题、数字、菜单文字 |
| `--bs-body-color` | `#646E78` | 正文 |
| `--bs-secondary-color` | `#A7ACB2` | 辅助文字、占位符 |
| `--bs-body-bg` | `#F5F5F9` | 页面背景 |
| `--bs-paper-bg` | `#FFFFFF` | 卡片、菜单、navbar |
| `--bs-border-color` | `#E4E6E8` | 边框、分隔线 |
| `--bs-border-color-translucent` | `rgba(34,48,62,.175)` | 半透明边框 |

状态浅色背景：primary `#E7E7FF`、secondary `#EBEEF0`、success `#E8FADF`、info `#D7F5FC`、warning `#FFF2D6`、danger `#FFE0DB`。

## 字体与尺寸

```css
--bs-font-sans-serif: "Public Sans", -apple-system, blinkmacsystemfont, "Segoe UI", Oxygen,
  "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
--bs-body-font-size: .9375rem; /* 15px */
--bs-body-font-weight: 400;
--bs-body-line-height: 1.375; /* 20.625px */
```

默认标题：`h1 2.875rem`、`h2 2.375rem`、`h3 1.75rem`、`h4 1.5rem`、`h5 1.125rem`、`h6 .9375rem`，标题字重 500、行高 1.1。`.small`/辅助文字约 13px；按钮通常沿用 15px，字重由按钮变体决定。字间距未发现统一自定义 token，视为 `normal`（待确认）。

## 圆角、阴影、透明度

| 令牌 | 值 |
|---|---|
| `--bs-border-radius-sm` | 4px |
| `--bs-border-radius` | 6px |
| `--bs-border-radius-lg` | 8px |
| `--bs-border-radius-xl` | 10px |
| `--bs-border-radius-xxl` | 16px |
| `--bs-border-radius-pill` | 50rem |
| `--bs-box-shadow-sm` | `0 2px 6px rgba(34,48,62,.08)` |
| `--bs-box-shadow` | `0 3px 8px rgba(34,48,62,.10)` |
| `--bs-box-shadow-lg` | `0 4px 12px rgba(34,48,62,.14)` |
| `--bs-box-shadow-inset` | `inset 0 1px 2px rgba(34,48,62,.075)` |
| focus ring | 宽 `0.15rem`，透明度 `.75` |

### 间距与断点

Bootstrap 间距基准：0、2px、4px、6px、8px、12px、16px、20px、24px、28px、32px、36px、40px、44px、48px。默认 `--bs-gutter-x: 1.625rem`（26px），`--bs-gutter-y: 0`。

断点：`xs:0`、`sm:576px`、`md:768px`、`lg:992px`、`xl:1200px`、`xxl:1400px`。
