
import { useEffect, useMemo, useState } from 'react';
import Button from '@mui/material/Button';
import RefreshOutlinedIcon from '@mui/icons-material/RefreshOutlined';
import VerifiedOutlinedIcon from '@mui/icons-material/VerifiedOutlined';
import { listAudits } from '../api/audit';
import type { AuditLog } from '../types/domain';
import { formatDate } from '../utils/format';
import { usePolling } from '../hooks/usePolling';
export default function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [selected, setSelected] = useState<AuditLog | null>(null);
  const [error, setError] = useState('');
  const load = async () => { try { const result = await listAudits(1, 100); setLogs(result.data); setError(''); } catch (reason) { setError(reason instanceof Error ? reason.message : String(reason)); } };
  useEffect(() => { void load(); }, []);
  usePolling(load, 20000);
  const transitions = useMemo(() => logs.filter((log) => log.action === 'transition').length, [logs]);
  const actors = useMemo(() => new Set(logs.map((log) => log.actor)).size, [logs]);
  return <main className="workspace"><header className="page-header"><div><p className="eyebrow">AUDIT GOVERNANCE</p><h1>审计追踪</h1><p>记录操作者、request ID、传感器证据和所有状态迁移。</p></div><Button startIcon={<RefreshOutlinedIcon />} onClick={() => void load()}>刷新</Button></header>
    <section className="metrics"><section className="metric"><span>审计记录</span><strong>{logs.length}</strong><small>当前查询范围</small></section><section className="metric"><span>状态迁移</span><strong>{transitions}</strong><small>不可覆盖</small></section><section className="metric"><span>操作人员</span><strong>{actors}</strong><small>责任身份</small></section></section>
    <section className="audit-integrity"><VerifiedOutlinedIcon /><span><strong>审计链完整</strong>状态迁移包含前态、后态、证据、操作者和 request ID。</span></section>
    {error && <div className="alert" role="alert">{error}</div>}<section className="table-shell"><table><thead><tr><th>时间</th><th>操作人</th><th>动作</th><th>对象</th><th>状态变化</th><th>request ID</th><th></th></tr></thead><tbody>{logs.map((log) => <tr key={log.id}><td>{formatDate(log.createdAt)}</td><td><strong>{log.actor}</strong></td><td><code>{log.action}</code></td><td>{log.entityType} #{log.entityId}</td><td>{log.beforeState || '-'} → {log.afterState || '-'}</td><td><small>{log.requestId}</small></td><td><button className="table-action" onClick={() => setSelected(selected?.id === log.id ? null : log)}>查看</button></td></tr>)}</tbody></table>{selected && <pre className="audit-detail">{JSON.stringify({ actor: selected.actor, action: selected.action, entity: `${selected.entityType} #${selected.entityId}`, before: selected.beforeState, after: selected.afterState, requestId: selected.requestId, detail: (() => { try { return JSON.parse(selected.detail); } catch { return selected.detail; } })() }, null, 2)}</pre>}</section>
  </main>;
}
