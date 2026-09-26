
import AttachFileOutlinedIcon from '@mui/icons-material/AttachFileOutlined';
export function EvidenceList({ evidence }: { evidence?: string | string[] }) {
  const items = Array.isArray(evidence) ? evidence : (evidence || '').split(',').map((item) => item.trim()).filter(Boolean);
  if (!items.length) return <div className="evidence-empty">暂无传感器证据</div>;
  return <div className="evidence-list">{items.map((item) => <span key={item}><AttachFileOutlinedIcon />{item}</span>)}</div>;
}
