
import { useEffect, useMemo, useState } from 'react';
import Button from '@mui/material/Button';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import SyncOutlinedIcon from '@mui/icons-material/SyncOutlined';
import { useTransportContainerStore } from '../stores/transport-container';
import type { DomainRecord } from '../types/domain';
import { roleAtLeast } from '../types/domain';
import { getSession } from '../api/client';
import { TemperatureBadge } from '../components/common/TemperatureBadge';
import { StatusBadge } from '../components/common/StatusBadge';
import { EvidenceList } from '../components/common/EvidenceList';
import { MetricCard } from '../components/common/MetricCard';
import { ConfirmDialog } from '../components/common/ConfirmDialog';
import { formatDate } from '../utils/format';

function nextContainerState(item: DomainRecord, canReview: boolean) {
  if (item.status === 'ready') return 'in_transit';
  if (item.status === 'in_transit') return 'quarantine';
  if (item.status === 'quarantine' && canReview) return 'cleared';
  if (item.status === 'cleared' && canReview) return 'quarantine';
  return '';
}

function temperatureRange(item: DomainRecord) {
  return (item.containerType || item.category).includes('干冰') ? { minimum: -80, maximum: -60 } : { minimum: 2, maximum: 8 };
}

export default function TransportContainerPage() {
  const store = useTransportContainerStore();
  const [search, setSearch] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [pending, setPending] = useState<{ item: DomainRecord; state: string } | null>(null);
  const session = getSession();
  const canOperate = roleAtLeast(session?.role, 'operator');
  const canReview = roleAtLeast(session?.role, 'reviewer');
  useEffect(() => { void store.load('containers'); }, [store.load]);
  const quarantine = useMemo(() => store.items.filter((item) => item.status === 'quarantine').length, [store.items]);
  const current = useMemo(() => store.items.filter((item) => item.status === 'in_transit').length, [store.items]);
  const createContainer = async () => {
    const suffix = Date.now().toString().slice(-5);
    const now = new Date().toISOString();
    await store.createRecord('containers', { code: `TC-UI-${suffix}`, name: '临床样本应急箱', description: '由质量工作台登记的应急运输容器', facility: '上海配送中心', owner: session?.username || 'operator', category: '主动制冷箱', riskLevel: 'low', metricValue: 4.5, metricUnit: 'C', effectiveAt: now, evidence: `minio://sensor/tc-ui-${suffix}/baseline.json`, relatedCode: `SN-UI-${suffix}`, sensorId: `SN-UI-${suffix}`, containerType: '主动制冷箱', currentLocation: '上海配送中心', custodian: session?.displayName || '现场操作员', currentTempC: 4.5, lastSensorReading: now });
    setCreateOpen(false);
  };
  const confirmTransition = async () => {
    if (!pending) return;
    await store.transition('containers', pending.item, pending.state, pending.state === 'quarantine' ? '温度异常，立即隔离容器' : '现场复核确认状态迁移', pending.item.evidence);
    setPending(null);
  };
  return <main className="workspace"><header className="page-header"><div><p className="eyebrow">CONTAINER MONITORING</p><h1>运输容器</h1><p>按传感器追踪当前位置、实时温度和隔离状态。</p></div>{canOperate && <Button variant="contained" startIcon={<AddOutlinedIcon />} onClick={() => setCreateOpen(true)}>登记容器</Button>}</header>
    <section className="metrics"><MetricCard label="在册容器" value={store.meta.total} detail="传感器已绑定" /><MetricCard label="运输中" value={current} detail="持续轮询温度" /><MetricCard label="隔离" value={quarantine} detail="禁止进入下游" /></section>
    <section className="toolbar"><input aria-label="搜索容器" placeholder="容器编码、名称" value={search} onChange={(event) => setSearch(event.target.value)} /><Button startIcon={<SyncOutlinedIcon />} onClick={() => void store.load('containers', search)}>查询</Button></section>
    {store.error && <div className="alert" role="alert">{store.error}</div>}
    <section className="table-shell" aria-busy={store.loading}><table><thead><tr><th>容器 / 传感器</th><th>位置</th><th>当前温度</th><th>状态</th><th>保管人</th><th>最近读数</th><th>证据</th><th>操作</th></tr></thead><tbody>{store.items.map((item) => { const next = nextContainerState(item, canReview); const range = temperatureRange(item); return <tr key={item.id}><td><strong>{item.code}</strong><small>{item.sensorId || item.relatedCode} · {item.containerType || item.category}</small></td><td>{item.currentLocation || item.facility}</td><td><TemperatureBadge value={item.currentTempC ?? item.metricValue} minimum={range.minimum} maximum={range.maximum} /></td><td><StatusBadge status={item.status} /></td><td>{item.custodian || item.owner}</td><td>{formatDate(item.lastSensorReading || item.updatedAt)}</td><td><EvidenceList evidence={item.evidence} /></td><td>{canOperate && next ? <button className="table-action" onClick={() => setPending({ item, state: next })}>{next === 'quarantine' ? '隔离' : next === 'cleared' ? '质量放行' : '开始运输'}</button> : <span className="muted">只读</span>}</td></tr>; })}</tbody></table>{store.loading && <div className="loading">正在同步传感器…</div>}</section>
    <ConfirmDialog open={createOpen} title="登记应急运输容器" onCancel={() => setCreateOpen(false)} onConfirm={() => void createContainer()}><p>将绑定新传感器并保存初始温度基线。</p></ConfirmDialog>
    <ConfirmDialog open={Boolean(pending)} title="确认容器状态" onCancel={() => setPending(null)} onConfirm={() => void confirmTransition()}><p>状态变化会写入审计记录；隔离后必须由质量复核员放行。</p><strong>{pending?.item.code}：{pending?.item.status} → {pending?.state}</strong></ConfirmDialog>
  </main>;
}
