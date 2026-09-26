
import type { ReactNode } from 'react';
import { UiButton } from './UiButton';
export function ConfirmDialog({ open, title, children, onConfirm, onCancel }: { open: boolean; title: string; children: ReactNode; onConfirm: () => void; onCancel: () => void }) {
  if (!open) return null;
  return <div className="modal-backdrop"><section className="modal" role="dialog" aria-modal="true" aria-label={title}><h2>{title}</h2><div>{children}</div><footer><button className="link-button" onClick={onCancel}>取消</button><UiButton onClick={onConfirm}>确认</UiButton></footer></section></div>;
}
