import { useAuth } from "../auth/AuthProvider.jsx";
import { PageHeader } from "../../app/components.jsx";

export function ProfilePage() { const { state } = useAuth(); return <><PageHeader code="SESSION" title="当前会话" description="查看当前浏览器会话的基本信息。" /><section className="card detail">{[["用户", state.displayName || state.nickname || state.subject], ["Subject", state.subject], ["Email", state.email || "未获取"], ["Application", state.applicationCode], ["有效期", state.expiresAt]].map(([label, value]) => <div className="detail-row" key={label}><span>{label}</span><strong>{value}</strong></div>)}<div className="permissions"><span>当前权限</span><div>{state.permissions?.map((item) => <code key={item}>{item}</code>)}</div></div></section></>; }
