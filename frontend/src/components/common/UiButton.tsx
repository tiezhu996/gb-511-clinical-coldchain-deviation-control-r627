import type { ReactNode } from 'react';
import Button from '@mui/material/Button';
export function UiButton({ children, onClick, disabled = false, danger = false }: { children: ReactNode; onClick?: () => void; disabled?: boolean; danger?: boolean }) {
  return <Button variant="contained" color={danger ? "error" : "primary"} onClick={onClick} disabled={disabled}>{children}</Button>;
}
