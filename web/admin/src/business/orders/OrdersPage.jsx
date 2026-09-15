import { useEffect, useMemo, useState } from "react";
import { request } from "../../modules/auth/api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";

export function OrdersPage() {
  const [data, setData] = useState(null); const [error, setError] = useState(null); const [query, setQuery] = useState(""); const [status, setStatus] = useState("all"); const [searchOpen, setSearchOpen] = useState(false);
  const load = () => request("/v1/orders").then((value) => setData(value.data || [])).catch(setError);
  useEffect(() => { load(); }, []);
  const filtered = useMemo(() => (data || []).filter((item) => (!query || item.id.toLowerCase().includes(query.toLowerCase())) && (status === "all" || item.status === status)), [data, query, status]);
  const suggestions = useMemo(() => (query ? filtered : data || []).slice(0, 5), [data, filtered, query]);
  const count = (value) => (data || []).filter((item) => item.status === value).length;
  if (error) return <><PageHeader code="ORDERS" title="订单查询" description="按订单号和状态筛选当前应用可访问的订单。" /><ErrorState error={error} retry={() => { setError(null); load(); }} /></>;
  if (!data) return <Loading text="正在加载订单…" />;
  return <>
    <PageHeader code="ECOMMERCE / ORDERS" title="订单查询" description="按订单号和状态筛选当前应用可访问的订单。" />
    <section className="order-stat-grid">
      {[['全部订单', data.length, 'all'], ['待支付', count('pending'), 'pending'], ['已完成', count('completed'), 'completed'], ['已退款', count('refunded'), 'refunded']].map(([label, value, key]) => <button className={`card order-stat ${status === key ? 'selected' : ''}`} key={key} type="button" onClick={() => setStatus(key)}><strong>{value}</strong><span>{label}</span></button>)}
    </section>
    <section className="card order-list-card">
      <div className="order-toolbar"><div className={`order-search-box ${searchOpen ? 'is-open' : ''}`} onBlur={(event) => { if (!event.currentTarget.contains(event.relatedTarget)) setSearchOpen(false); }}>
        <label className="order-search"><span className="sr-only">搜索订单号</span><svg aria-hidden="true" viewBox="0 0 24 24"><circle cx="10.8" cy="10.8" r="6.8"/><path d="m16 16 4.2 4.2"/></svg><input value={query} onFocus={() => setSearchOpen(true)} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => { if (event.key === 'Escape') setSearchOpen(false); }} placeholder="搜索订单号…" aria-expanded={searchOpen} aria-controls="order-search-results" aria-autocomplete="list" /></label>
        {searchOpen && <div className="order-search-results" id="order-search-results" role="listbox" aria-label="订单搜索结果"><div className="order-search-heading">订单</div>{suggestions.length ? suggestions.map((item) => <button className="order-search-result" key={item.id} type="button" role="option" aria-selected="false" onClick={() => { setQuery(item.id); setSearchOpen(false); }}><span className="order-search-result-icon">⌕</span><span className="order-search-result-copy"><strong>{item.id}</strong><small>{item.createdAt || '订单'} · {item.status}</small></span><span className="order-search-result-arrow">↵</span></button>) : <div className="order-search-empty">没有找到匹配的订单</div>}</div>}
      </div><select aria-label="订单状态" value={status} onChange={(event) => setStatus(event.target.value)}><option value="all">全部状态</option><option value="pending">待支付</option><option value="completed">已完成</option><option value="refunded">已退款</option><option value="failed">失败</option></select><button type="button" className="text-button" onClick={load}>刷新</button></div>
      {filtered.length === 0 ? <div className="empty"><strong>{data.length ? '没有匹配订单' : '暂无订单'}</strong><span>{data.length ? '请调整搜索条件或状态筛选。' : '当前没有可展示的订单记录。'}</span></div> : <div className="table-wrap"><table><thead><tr><th>订单编号</th><th>创建时间</th><th>客户</th><th>支付</th><th>状态</th><th>金额</th></tr></thead><tbody>{filtered.map((item) => <tr key={item.id}><td className="mono">{item.id}</td><td>{item.createdAt}</td><td>—</td><td><span className="tag">{item.status}</span></td><td><span className="tag">{item.status}</span></td><td>{(item.totalMinor / 100).toFixed(2)} {item.currency}</td></tr>)}</tbody></table></div>}
    </section>
    <style>{`
      .order-search-box { position: relative; flex: 1; min-width: 12rem; }
      .order-search { position: relative; display: block; }
      .order-search svg { position: absolute; z-index: 1; top: 50%; left: .85rem; width: 1.1rem; height: 1.1rem; fill: none; stroke: #a1acb8; stroke-width: 1.7; transform: translateY(-50%); pointer-events: none; }
      .order-search input { padding: .68rem 2.5rem .68rem 2.55rem; outline: 0; transition: border-color 140ms ease, box-shadow 140ms ease; }
      .order-search input::placeholder { color: #a1acb8; }
      .order-search-box.is-open .order-search input { border-color: #696cff; box-shadow: 0 0 0 .15rem rgba(105,108,255,.13); }
      .order-search-results { position: absolute; z-index: 30; top: calc(100% + .45rem); left: 0; right: 0; padding: .55rem 0; overflow: hidden; border: 1px solid #e4e6e8; border-radius: .5rem; background: #fff; box-shadow: 0 .5rem 1.25rem rgba(34,48,62,.16); }
      .order-search-heading { padding: .35rem 1rem .45rem; color: #a1acb8; font-size: .68rem; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
      .order-search-result { display: flex; align-items: center; gap: .75rem; width: 100%; padding: .65rem 1rem; color: #566a7f; background: #fff; cursor: pointer; text-align: left; }
      .order-search-result:hover, .order-search-result:focus-visible { outline: 0; background: #f5f5f9; }
      .order-search-result-icon { display: grid; width: 2rem; height: 2rem; flex: 0 0 2rem; place-items: center; border-radius: .375rem; color: #696cff; background: rgba(105,108,255,.12); font-size: 1.1rem; }
      .order-search-result-copy { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: .15rem; }
      .order-search-result-copy strong { overflow: hidden; color: #384551; font-size: .83rem; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
      .order-search-result-copy small { overflow: hidden; color: #a1acb8; font-size: .72rem; text-overflow: ellipsis; white-space: nowrap; }
      .order-search-result-arrow { color: #a1acb8; font-size: .82rem; }
      .order-search-empty { padding: .85rem 1rem; color: #8592a3; font-size: .82rem; }
      @media (max-width: 700px) { .order-toolbar { align-items: stretch; flex-wrap: wrap; } .order-search-box { flex-basis: 100%; } .order-toolbar select { flex: 1; width: auto; } }
    `}</style>
  </>;
}
