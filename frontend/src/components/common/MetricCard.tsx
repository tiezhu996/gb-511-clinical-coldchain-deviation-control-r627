
import type { ReactNode } from 'react';
export function MetricCard({ label, value, detail }: { label: string; value: ReactNode; detail?: string }) {
  return <section className="metric"><span>{label}</span><strong>{value}</strong>{detail && <small>{detail}</small>}</section>;
}
