import GavelOutlinedIcon from '@mui/icons-material/GavelOutlined';
import type { DomainRecord } from '../../types/domain';
import { StatusBadge } from './StatusBadge';
import { EvidenceList } from './EvidenceList';

export function DecisionPanel({ decision, compact = false }: { decision?: DomainRecord | null; compact?: boolean }) {
  return <section className={`decision-panel ${compact ? 'decision-panel--compact' : ''}`}>
    <header><span><GavelOutlinedIcon />处置决定</span>{decision ? <StatusBadge status={decision.status} /> : <StatusBadge status="pending" />}</header>
    {decision ? <><p>{decision.decisionBasis || decision.description}</p><dl><div><dt>偏差事件</dt><dd>{decision.excursionCode || decision.relatedCode}</dd></div><div><dt>提议人</dt><dd>{decision.proposedBy || decision.owner}</dd></div><div><dt>复核人</dt><dd>{decision.approvedBy || '待独立复核'}</dd></div></dl><EvidenceList evidence={decision.sensorEvidence || decision.evidence} /></> : <p className="muted">尚未形成处置提议。</p>}
  </section>;
}
