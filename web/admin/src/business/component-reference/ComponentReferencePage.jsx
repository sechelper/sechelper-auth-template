import { useMemo, useState } from "react";
import "./component-reference.css";

const rows = [
  { id: "ORD-24018", customer: "林知夏", date: "2026-09-14", amount: 1280, status: "已完成" },
  { id: "ORD-24017", customer: "周予安", date: "2026-09-14", amount: 368, status: "处理中" },
  { id: "ORD-24016", customer: "陈星野", date: "2026-09-13", amount: 89, status: "待确认" },
  { id: "ORD-24015", customer: "许清禾", date: "2026-09-13", amount: 720, status: "已取消" },
  { id: "ORD-24014", customer: "沈知遥", date: "2026-09-12", amount: 245, status: "已完成" },
  { id: "ORD-24013", customer: "顾言川", date: "2026-09-12", amount: 1560, status: "处理中" },
  { id: "ORD-24012", customer: "唐晚晴", date: "2026-09-11", amount: 98, status: "待确认" },
  { id: "ORD-24011", customer: "陆景行", date: "2026-09-11", amount: 432, status: "已完成" },
  { id: "ORD-24010", customer: "苏念慈", date: "2026-09-10", amount: 850, status: "已取消" },
  { id: "ORD-24009", customer: "江予白", date: "2026-09-10", amount: 210, status: "处理中" },
  { id: "ORD-24008", customer: "叶知秋", date: "2026-09-09", amount: 640, status: "已完成" },
  { id: "ORD-24007", customer: "温以宁", date: "2026-09-09", amount: 315, status: "待确认" },
];

const icons = [
  ["home", "导航", "⌂"], ["dashboard", "导航", "▦"], ["menu", "导航", "☰"], ["search", "操作", "⌕"],
  ["plus", "操作", "+"], ["edit", "操作", "✎"], ["trash", "操作", "⌫"], ["download", "操作", "↓"],
  ["upload", "操作", "↑"], ["filter", "操作", "▽"], ["sort", "操作", "↕"], ["refresh", "操作", "↻"],
  ["user", "用户", "♙"], ["users", "用户", "♧"], ["settings", "系统", "⚙"], ["bell", "系统", "♧"],
  ["calendar", "日期", "▦"], ["clock", "日期", "◷"], ["check", "状态", "✓"], ["close", "状态", "×"],
  ["info", "状态", "i"], ["warning", "状态", "!"], ["error", "状态", "×"], ["help", "状态", "?"],
  ["folder", "文件", "▱"], ["file", "文件", "▤"], ["image", "文件", "▧"], ["link", "通用", "↗"],
  ["lock", "安全", "▣"], ["shield", "安全", "◇"], ["chart", "数据", "◒"], ["table", "数据", "▦"],
];

const sections = ["总览", "表格", "图标", "表单", "按钮与操作", "反馈与状态", "布局与导航", "Frest 组件目录"];

const templateComponents = [
  ["卡片", "基础卡片", "cards/basic"], ["卡片", "高级卡片", "cards/advance"], ["卡片", "统计卡片", "cards/statistics"], ["卡片", "分析卡片", "cards/analytics"], ["卡片", "操作卡片", "cards/actions"],
  ["用户界面", "手风琴", "ui/accordion"], ["用户界面", "提示信息（Alerts）", "ui/alerts"], ["用户界面", "徽章", "ui/badges"], ["用户界面", "按钮", "ui/buttons"], ["用户界面", "轮播", "ui/carousel"], ["用户界面", "折叠面板", "ui/collapse"], ["用户界面", "下拉菜单", "ui/dropdowns"], ["用户界面", "页脚", "ui/footer"], ["用户界面", "列表组", "ui/list-groups"], ["用户界面", "模态框", "ui/modals"], ["用户界面", "导航栏", "ui/navbar"], ["用户界面", "侧边抽屉（Offcanvas）", "ui/offcanvas"], ["用户界面", "分页与面包屑", "ui/pagination-breadcrumbs"], ["用户界面", "进度条", "ui/progress"], ["用户界面", "加载动画", "ui/spinners"], ["用户界面", "标签页与 Pills", "ui/tabs-pills"], ["用户界面", "Toast 消息", "ui/toasts"], ["用户界面", "工具提示与 Popovers", "ui/tooltips-popovers"], ["用户界面", "排版", "ui/typography"],
  ["扩展界面", "头像", "extended/ui-avatar"], ["扩展界面", "BlockUI 页面遮罩", "extended/ui-blockui"], ["扩展界面", "拖放", "extended/ui-drag-and-drop"], ["扩展界面", "媒体播放器", "extended/ui-media-player"], ["扩展界面", "滚动条", "extended/ui-perfect-scrollbar"], ["扩展界面", "星级评分", "extended/ui-star-ratings"], ["扩展界面", "SweetAlert2 弹窗", "extended/ui-sweetalert2"], ["扩展界面", "文字分隔线", "extended/ui-text-divider"], ["扩展界面", "基础时间线", "extended/ui-timeline-basic"], ["扩展界面", "全屏时间线", "extended/ui-timeline-fullscreen"], ["扩展界面", "页面导览（Tour）", "extended/ui-tour"], ["扩展界面", "树形视图", "extended/ui-treeview"], ["扩展界面", "其他扩展组件", "extended/ui-misc"],
  ["图标", "Boxicons", "icons/boxicons"], ["图标", "Font Awesome", "icons/font-awesome"],
  ["表单控件", "基础输入框", "forms/basic-inputs"], ["表单控件", "输入组", "forms/input-groups"], ["表单控件", "自定义选项", "forms/custom-options"], ["表单控件", "富文本编辑器", "forms/editors"], ["表单控件", "文件上传", "forms/file-upload"], ["表单控件", "日期与时间选择器", "forms/pickers"], ["表单控件", "下拉选择与标签", "forms/selects"], ["表单控件", "滑块", "forms/sliders"], ["表单控件", "开关", "forms/switches"], ["表单控件", "其他表单控件", "forms/extras"],
  ["表单布局与验证", "垂直表单", "form/layouts-vertical"], ["表单布局与验证", "水平表单", "form/layouts-horizontal"], ["表单布局与验证", "固定操作区", "form/layouts-sticky"], ["表单布局与验证", "数字步骤向导", "form/wizard-numbered"], ["表单布局与验证", "图标步骤向导", "form/wizard-icons"], ["表单布局与验证", "表单验证", "form/validation"],
  ["表格", "基础表格", "tables/basic"], ["表格", "基础数据表格", "tables/datatables-basic"], ["表格", "高级数据表格", "tables/datatables-advanced"], ["表格", "数据表格扩展", "tables/datatables-extensions"],
  ["图表与地图", "ApexCharts", "charts/apex"], ["图表与地图", "Chart.js", "charts/chartjs"], ["图表与地图", "Leaflet 地图", "maps/leaflet"],
];

function Label({ children, tone = "neutral" }) { return <span className={`ref-label ref-label-${tone}`}>{children}</span>; }
function SectionTitle({ id, title, description }) { return <div className="ref-section-title" id={id}><div><h2>{title}</h2><p>{description}</p></div><a href="#reference-top">返回顶部 ↑</a></div>; }
function Icon({ item }) { return <span className="ref-icon-glyph" aria-hidden="true">{item[2]}</span>; }

export function ComponentReferencePage() {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("全部状态");
  const [sort, setSort] = useState("date");
  const [ascending, setAscending] = useState(false);
  const [page, setPage] = useState(1);
  const [empty, setEmpty] = useState(false);
  const [tableLoading, setTableLoading] = useState(false);
  const [iconQuery, setIconQuery] = useState("");
  const [iconCategory, setIconCategory] = useState("全部");
  const [form, setForm] = useState({ name: "", email: "", quantity: "2", category: "标准", tags: ["后台"], role: "viewer", enabled: true, date: "2026-09-15", note: "示例内容，可编辑。" });
  const [formError, setFormError] = useState("");
  const [formBusy, setFormBusy] = useState(false);
  const [formDone, setFormDone] = useState(false);
  const [toast, setToast] = useState("");
  const [confirming, setConfirming] = useState(false);
  const [tab, setTab] = useState("概要");
  const [componentQuery, setComponentQuery] = useState("");
  const [componentCategory, setComponentCategory] = useState("全部");

  const filteredRows = useMemo(() => {
    const result = (empty ? [] : rows).filter((row) => (!query || `${row.id} ${row.customer}`.toLowerCase().includes(query.toLowerCase())) && (status === "全部状态" || row.status === status));
    result.sort((a, b) => String(a[sort]).localeCompare(String(b[sort]), "zh-CN", { numeric: true }) * (ascending ? 1 : -1));
    return result;
  }, [empty, query, status, sort, ascending]);
  const iconCategories = ["全部", ...new Set(icons.map((item) => item[1]))];
  const visibleIcons = icons.filter((item) => (iconCategory === "全部" || item[1] === iconCategory) && (!iconQuery || `${item[0]} ${item[1]}`.toLowerCase().includes(iconQuery.toLowerCase())));
  const componentCategories = ["全部", ...new Set(templateComponents.map(([category]) => category))];
  const visibleComponents = templateComponents.filter(([category, name]) => (componentCategory === "全部" || category === componentCategory) && (!componentQuery || `${category} ${name}`.toLowerCase().includes(componentQuery.toLowerCase())));
  const rowsPerPage = 5;
  const pageCount = Math.max(1, Math.ceil(filteredRows.length / rowsPerPage));
  const pageRows = filteredRows.slice((page - 1) * rowsPerPage, page * rowsPerPage);
  const update = (key, value) => { setForm((current) => ({ ...current, [key]: value })); setFormDone(false); };

  function submitForm(event) {
    event.preventDefault();
    if (!form.name.trim()) { setFormError("请输入名称，这是必填字段。"); return; }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) { setFormError("请输入有效的邮箱地址。"); return; }
    setFormError(""); setFormBusy(true); setFormDone(false);
    window.setTimeout(() => { setFormBusy(false); setFormDone(true); }, 550);
  }

  function notify(message) { setToast(message); window.setTimeout(() => setToast(""), 2400); }

  return <main className="ref-page" id="reference-top">
    <div className="ref-page-heading">
      <div><div className="ref-eyebrow">DEVELOPER TOOLKIT <span>TEST ONLY</span></div><h1>前端组件参考</h1><p>统一管理端常见组件的结构、字段、状态和交互。示例均为虚构数据，只在开发/测试构建中出现，不连接业务写接口。</p></div>
      <div className="ref-build-stamp"><strong>仅开发 / 测试构建</strong><small>PRODUCTION BUNDLE EXCLUDES THIS PAGE</small></div>
    </div>
    <nav className="ref-toc" aria-label="组件目录">{sections.map((name, index) => <a key={name} href={`#ref-${index}`}>{String(index + 1).padStart(2, "0")}　{name}</a>)}</nav>

    <section className="ref-overview-grid" id="ref-0">
      <article className="ref-card"><span className="ref-overview-icon">⌘</span><strong>沿用现有后台</strong><p>色彩、卡片、表格密度与 Sneat/Frest 风格管理壳一致，不另造全局设计系统。</p></article>
      <article className="ref-card"><span className="ref-overview-icon">◎</span><strong>可操作样例</strong><p>筛选、排序、分页、表单校验、标签页和确认反馈均可在本页体验。</p></article>
      <article className="ref-card"><span className="ref-overview-icon">⌁</span><strong>接口对齐提示</strong><p>每个示例说明建议字段、状态值与行为；前端交互不代表服务端授权或校验。</p></article>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-1" title="表格" description="列表接口建议提供稳定主键、分页游标/页码、可排序字段和明确的状态枚举。以下交互仅作用于本地虚构数据。" />
      <div className="ref-toolbar"><input aria-label="搜索订单号或客户" placeholder="搜索编号或客户…" value={query} onChange={(event) => { setQuery(event.target.value); setPage(1); setEmpty(false); }} /><select aria-label="按状态筛选" value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}><option>全部状态</option><option>已完成</option><option>处理中</option><option>待确认</option><option>已取消</option></select><button className="ref-button ref-button-outline" type="button" onClick={() => { setTableLoading(true); window.setTimeout(() => setTableLoading(false), 450); }}>↻ 刷新示例</button><button className="ref-button ref-button-quiet" type="button" onClick={() => { setEmpty(!empty); setPage(1); }}>{empty ? "显示数据" : "预览空状态"}</button></div>
      <div className="ref-table-wrap"><table className="ref-table"><thead><tr>{[["编号", "id"], ["客户", "customer"], ["创建日期", "date"], ["金额", "amount"], ["状态", "status"]].map(([label, field]) => <th key={field}><button className="ref-sort" type="button" onClick={() => { setSort(field); setAscending(sort === field ? !ascending : true); }}>{label} <span>{sort === field ? (ascending ? "↑" : "↓") : "↕"}</span></button></th>)}<th>行操作</th></tr></thead><tbody>{pageRows.map((row) => <tr key={row.id}><td className="ref-mono">{row.id}</td><td>{row.customer}</td><td>{row.date}</td><td>¥ {row.amount.toLocaleString("zh-CN", { minimumFractionDigits: 2 })}</td><td><Label tone={row.status === "已完成" ? "success" : row.status === "处理中" ? "primary" : row.status === "已取消" ? "danger" : "warning"}>{row.status}</Label></td><td><button className="ref-link-button" type="button" onClick={() => notify(`查看 ${row.id}（演示）`)}>查看</button><button className="ref-link-button" type="button" onClick={() => setConfirming(true)}>更多</button></td></tr>)}</tbody></table>
        {tableLoading && <div className="ref-table-overlay"><span className="ref-spinner" />正在载入示例…</div>}
        {!tableLoading && pageRows.length === 0 && <div className="ref-empty"><span>⌕</span><strong>{empty ? "暂无数据" : "没有匹配结果"}</strong><small>{empty ? "接口返回空数组时展示此状态。" : "调整搜索词或筛选条件后重试。"}</small></div>}
      </div>
      <div className="ref-table-footer"><small>共 {filteredRows.length} 条虚构记录 · 第 {page} / {pageCount} 页</small><div><button className="ref-page-button" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</button><button className="ref-page-button" disabled={page >= pageCount} onClick={() => setPage(page + 1)}>下一页</button></div></div>
      <p className="ref-note"><strong>字段参考：</strong><code>id</code>（稳定主键）、<code>createdAt</code>（ISO 日期时间）、<code>amount</code>（最小货币单位或 decimal 字符串）、<code>status</code>（服务端枚举）。排序/筛选条件提交给后端时应做白名单和权限校验。</p>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-2" title="图标" description="使用项目自身的轻量符号示例，不依赖外部字体或复制第三方图标资源。点击名称可复制稳定的语义标识。" />
      <div className="ref-toolbar ref-icon-toolbar"><input aria-label="搜索图标" placeholder="搜索名称，例如 search…" value={iconQuery} onChange={(event) => setIconQuery(event.target.value)} /><div className="ref-chip-row">{iconCategories.map((category) => <button key={category} className={`ref-chip ${iconCategory === category ? "active" : ""}`} type="button" onClick={() => setIconCategory(category)}>{category}</button>)}</div></div>
      <div className="ref-icon-grid">{visibleIcons.map((item) => <button className="ref-icon-card" key={item[0]} type="button" onClick={() => { navigator.clipboard?.writeText(item[0]); notify(`已复制名称：${item[0]}`); }}><Icon item={item} /><span>{item[0]}</span><small>{item[1]}</small></button>)}</div>
      {visibleIcons.length === 0 && <div className="ref-empty ref-icon-empty"><strong>没有匹配的图标</strong><small>试试更短的名称或切换分类。</small></div>}
      <p className="ref-note">点击图标复制语义名称（如 <code>search</code>）。服务端字段无需存储图标字符；页面可将稳定枚举映射到本项目图标组件。</p>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-3" title="表单" description="字段名应与 API 请求 DTO 对齐；前端校验用于即时反馈，服务端仍需重复执行权威校验。" />
      <form className="ref-form" onSubmit={submitForm} noValidate>
        <label>名称 <em>*</em><input value={form.name} onChange={(event) => update("name", event.target.value)} placeholder="请输入名称" required aria-invalid={!!formError && !form.name} /><small>必填 · string · 最长 80 个字符</small></label>
        <label>邮箱 <em>*</em><input type="email" value={form.email} onChange={(event) => update("email", event.target.value)} placeholder="name@example.test" required /><small>必填 · email · 示例地址不会发送</small></label>
        <label>数量<input type="number" min="1" max="99" value={form.quantity} onChange={(event) => update("quantity", event.target.value)} /><small>number · min 1 · max 99</small></label>
        <label>分类<select value={form.category} onChange={(event) => update("category", event.target.value)}><option>标准</option><option>优先</option><option>归档</option></select><small>单选 · enum</small></label>
        <label>标签（多选）<select multiple value={form.tags} onChange={(event) => update("tags", [...event.target.selectedOptions].map((option) => option.value))}><option>后台</option><option>只读</option><option>待复核</option></select><small>多选 · string[]</small></label>
        <fieldset><legend>角色（单选）</legend>{["viewer", "editor", "owner"].map((role) => <label className="ref-choice" key={role}><input type="radio" name="reference-role" checked={form.role === role} onChange={() => update("role", role)} />{role}</label>)}<small>enum · 权限最终由服务端判断</small></fieldset>
        <label className="ref-switch-field">启用状态 <button className={`ref-switch ${form.enabled ? "on" : ""}`} type="button" role="switch" aria-checked={form.enabled} onClick={() => update("enabled", !form.enabled)}><span /></button><small>boolean · 开/关</small></label>
        <label>生效日期<input type="date" value={form.date} onChange={(event) => update("date", event.target.value)} /><small>date · YYYY-MM-DD</small></label>
        <label className="ref-span-two">备注<textarea rows="3" value={form.note} onChange={(event) => update("note", event.target.value)} placeholder="补充说明…" /><small>string · 最长 500 个字符</small></label>
        <label>禁用字段<input value="系统生成" disabled /><small>disabled · 不随表单提交</small></label>
        <label>只读字段<input value="REF-2026-001" readOnly /><small>readOnly · 值仍会随表单提交</small></label>
        <div className="ref-span-two ref-validation-hint">{formError && <span className="ref-error-text">{formError}</span>}{formDone && <span className="ref-success-text">示例校验通过；未发送请求或写入数据。</span>}</div>
        <div className="ref-span-two ref-form-actions"><button className="ref-button ref-button-primary" disabled={formBusy} type="submit">{formBusy && <span className="ref-spinner ref-spinner-light" />} {formBusy ? "校验中…" : "提交示例"}</button><button className="ref-button ref-button-outline" type="button" onClick={() => { setForm({ name: "", email: "", quantity: "1", category: "标准", tags: [], role: "viewer", enabled: false, date: "", note: "" }); setFormError(""); setFormDone(false); }}>重置</button><span className="ref-safe-note">仅前端模拟，不调用业务写接口</span></div>
      </form>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-4" title="按钮与操作" description="按钮文案描述动作；危险操作需明确影响并二次确认。加载态应避免重复提交。" />
      <div className="ref-demo-row"><button className="ref-button ref-button-primary" onClick={() => notify("主操作已触发（演示）")}>主按钮</button><button className="ref-button ref-button-outline" onClick={() => notify("次要操作已触发（演示）")}>次按钮</button><button className="ref-button ref-button-danger" onClick={() => setConfirming(true)}>危险操作</button><button className="ref-button ref-button-outline" disabled>禁用按钮</button><button className="ref-button ref-button-primary" disabled><span className="ref-spinner ref-spinner-light" /> 加载中</button></div>
      <div className="ref-demo-row"><button className="ref-button ref-button-primary ref-button-small">小号</button><button className="ref-button ref-button-primary">默认尺寸</button><button className="ref-button ref-button-primary ref-button-large">大号</button></div>
      <p className="ref-note">图标按钮应补充可访问名称；不可逆动作需要后端幂等/权限保护，前端确认框不构成安全边界。</p>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-5" title="反馈与状态" description="所有状态都用文字辅助颜色表达，方便识别并满足无障碍需求。" />
      <div className="ref-subsection"><h3>消息与确认</h3><div className="ref-demo-row"><button className="ref-button ref-button-primary" onClick={() => notify("操作已完成（本地示例）")}>显示提示消息</button><button className="ref-button ref-button-outline" onClick={() => setConfirming(true)}>打开确认框</button></div>{toast && <div role="status" className="ref-toast">✓　{toast}</div>}{confirming && <div className="ref-dialog-backdrop" role="presentation"><div className="ref-dialog" role="dialog" aria-modal="true" aria-labelledby="ref-dialog-title"><span className="ref-dialog-mark">!</span><h3 id="ref-dialog-title">确认此示例操作？</h3><p>这只是交互演示，不会修改任何数据。</p><div><button className="ref-button ref-button-outline" onClick={() => setConfirming(false)}>取消</button><button className="ref-button ref-button-danger" onClick={() => { setConfirming(false); notify("已确认示例操作"); }}>确认操作</button></div></div></div>}</div>
      <div className="ref-subsection"><h3>徽标、标签与进度</h3><div className="ref-demo-row"><Label tone="success">已完成</Label><Label tone="primary">进行中</Label><Label tone="warning">待处理</Label><Label tone="danger">失败</Label><span className="ref-count-badge">8</span><span className="ref-count-badge ref-count-danger">3</span></div><div className="ref-progress"><div><span>任务进度 <b>68%</b></span><div className="ref-progress-track"><i style={{ width: "68%" }} /></div></div><div><span>处理中 <b>32%</b></span><div className="ref-progress-track"><i className="ref-progress-green" style={{ width: "32%" }} /></div></div></div></div>
      <div className="ref-subsection"><h3>空状态与加载状态</h3><div className="ref-state-grid"><div className="ref-state-box"><span className="ref-state-symbol">⌕</span><strong>暂无记录</strong><small>接口返回空数组时展示。</small></div><div className="ref-state-box"><span className="ref-spinner" /><strong>正在加载</strong><small>避免无限等待；请求失败需展示重试。</small></div><div className="ref-state-box"><Label tone="danger">请求失败</Label><strong>数据暂不可用</strong><button className="ref-link-button" onClick={() => notify("重试请求（演示）")}>重试</button></div></div></div>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-6" title="布局与导航" description="列表页可通过面包屑说明层级；标签页和折叠面板用于组织同一任务中的相关内容。" />
      <div className="ref-breadcrumb"><a href="#reference-top">首页</a><span>/</span><a href="#ref-6">开发参考</a><span>/</span><strong>组件参考</strong></div>
      <div className="ref-tabs" role="tablist">{["概要", "字段说明", "变更记录"].map((item) => <button key={item} type="button" role="tab" aria-selected={tab === item} className={tab === item ? "active" : ""} onClick={() => setTab(item)}>{item}</button>)}</div>
      <div className="ref-tab-panel"><strong>{tab}</strong><p>{tab === "概要" ? "标签页切换内容，不应丢失未保存表单状态。" : tab === "字段说明" ? "字段名、类型、必填/可空和枚举值应与 API 合同保持一致。" : "最近更新：2026-09-15 · 此处为虚构展示。"}</p></div>
      <details className="ref-accordion" open><summary>折叠面板示例：请求与响应约定</summary><p>分页响应建议包含 <code>items</code>、<code>total</code> 或 <code>nextCursor</code>；时间统一明确时区；错误响应提供稳定错误码。</p></details>
      <details className="ref-accordion"><summary>折叠面板示例：授权提示</summary><p>前端根据权限展示操作入口，但每个业务接口仍必须在服务端执行最终鉴权。</p></details>
    </section>

    <section className="ref-card ref-component-section">
      <SectionTitle id="ref-7" title="Frest 组件目录" description="按原模板演示站的组件菜单整理，包含卡片、基础与扩展界面、图标、表单、表格、图表和地图。打开链接可查看对应的在线演示；目录用于开发选型，不复制模板源码。" />
      <div className="ref-toolbar ref-catalog-toolbar"><input aria-label="搜索模板组件" placeholder="搜索组件名称或分类…" value={componentQuery} onChange={(event) => setComponentQuery(event.target.value)} /><div className="ref-chip-row">{componentCategories.map((category) => <button key={category} className={`ref-chip ${componentCategory === category ? "active" : ""}`} type="button" onClick={() => setComponentCategory(category)}>{category}</button>)}</div></div>
      <div className="ref-catalog-grid">{visibleComponents.map(([category, name, path]) => <a className="ref-catalog-card" key={path} href={`https://demos.pixinvent.com/frest-html-laravel-admin-template/demo-1/${path}`} target="_blank" rel="noreferrer"><small>{category}</small><strong>{name}</strong><span>打开在线演示 ↗</span></a>)}</div>
      {visibleComponents.length === 0 && <div className="ref-empty ref-icon-empty"><strong>没有匹配的组件</strong><small>试试其他名称或切换分类。</small></div>}
      <p className="ref-note"><strong>目录范围：</strong>共 {templateComponents.length} 个演示入口。模板菜单另列有仪表盘、业务应用、账户页和认证页，它们属于页面示例，不计入此组件目录。</p>
    </section>
    <footer className="ref-page-footer">开发参考页 · 所有姓名、编号、状态和金额均为虚构数据 · 不构成正式业务功能</footer>
  </main>;
}
