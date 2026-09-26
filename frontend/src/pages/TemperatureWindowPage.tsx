
import { useEffect, useMemo, useState } from 'react';
import Button from '@mui/material/Button';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import PublishedWithChangesOutlinedIcon from '@mui/icons-material/PublishedWithChangesOutlined';
import { useTemperatureWindowStore } from '../stores/temperature-window';
import type { DomainRecord } from '../types/domain';
import { roleAtLeast } from '../types/domain';
import { getSession } from '../api/client';
import { StatusBadge } from '../components/common/StatusBadge';
import { MetricCard } from '../components/common/MetricCard';
import { EvidenceList } from '../components/common/EvidenceList';
import { ConfirmDialog } from '../components/common/ConfirmDialog';

export default function TemperatureWindowPage() {
  const store = useTemperatureWindowStore();
  const [pending, setPending] = useState<DomainRecord | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const session = getSession();
  const canReview = roleAtLeast(session?.role, 'reviewer');
  useEffect(() => { void store.load('windows'); }, [store.load]);
  const active = useMemo(() => store.items.filter((item) => item.status === 'active').length, [store.items]);
  const critical = useMemo(() => store.items.filter((item) => item.riskLevel === 'critical').length, [store.items]);
  const createWindow = async () => {
    const suffix = Date.now().toString().slice(-5); const now = new Date().toISOString();
    await store.createRecord('windows', { code: `TW-UI-${suffix}`, name: '临床样本 2-8C 规则草案', description: '等待质量负责人复核生效', facility: '质量体系 QMS-UI', owner: session?.username || 'reviewer', category: '临床样本', riskLevel: 'high', metricValue: 8, metricUnit: 'C', effectiveAt: now, evidence: 'DRAFT-SOP-UI', relatedCode: 'DRAFT-SOP-UI', productClass: '临床样本', minimumCelsius: 2, maximumCelsius: 8, maxExcursionMinutes: 15, qualityOwner: session?.username || 'reviewer' });
    setCreateOpen(false);
  };
  const activate = async () => { if (!pending) return; await store.transition('windows', pending, 'active', '质量负责人复核温控范围与允许时长', pending.evidence); setPending(null); };
  return <main className="workspace"><header className="page-header"><div><p className="eyebrow">TEMPERATURE POLICY</p><h1>温控规则</h1><p>维护产品温度窗口与可接受偏差时长，规则生效需质量角色确认。</p></div>{canReview && <Button variant="contained" startIcon={<AddOutlinedIcon />} onClick={() => setCreateOpen(true)}>新增规则</Button>}</header>
    <section className="metrics"><MetricCard label="规则总数" value={store.meta.total} detail="版本化管理" /><MetricCard label="生效中" value={active} detail="参与偏差判断" /><MetricCard label="关键规则" value={critical} detail="深低温或高风险" /></section>
    {store.error && <div className="alert" role="alert">{store.error}</div>}
    <section className="table-shell"><table><thead><tr><th>规则</th><th>产品类别</th><th>温度窗口</th><th>最长偏差</th><th>质量负责人</th><th>状态</th><th>依据</th><th>操作</th></tr></thead><tbody>{store.items.map((item) => <tr key={item.id}><td><strong>{item.code}</strong><small>{item.name}</small></td><td>{item.productClass || item.category}</td><td><strong>{item.minimumCelsius ?? 0} 至 {item.maximumCelsius ?? item.metricValue} C</strong></td><td>{item.maxExcursionMinutes || 0} 分钟</td><td>{item.qualityOwner || item.owner}</td><td><StatusBadge status={item.status} /></td><td><EvidenceList evidence={item.evidence} /></td><td>{canReview && item.status === 'draft' ? <Button size="small" startIcon={<PublishedWithChangesOutlinedIcon />} onClick={() => setPending(item)}>复核生效</Button> : <span className="muted">受控</span>}</td></tr>)}</tbody></table></section>
    <ConfirmDialog open={createOpen} title="创建温控规则草案" onCancel={() => setCreateOpen(false)} onConfirm={() => void createWindow()}><p>新规则默认保存为草稿，不会直接参与偏差判断。</p></ConfirmDialog>
    <ConfirmDialog open={Boolean(pending)} title="确认规则生效" onCancel={() => setPending(null)} onConfirm={() => void activate()}><p>请确认温度上下限、最长允许偏差时长和 SOP 依据已经复核。</p></ConfirmDialog>
  </main>;
}
