import type { EntityConfig } from './domain';

export type ContainerState = 'ready' | 'in_transit' | 'quarantine' | 'cleared';
export const ALL_CONTAINER_STATE: readonly ContainerState[] = ['ready', 'in_transit', 'quarantine', 'cleared'];
export type ExcursionState = 'open' | 'in_review' | 'decided' | 'closed';
export const ALL_EXCURSION_STATE: readonly ExcursionState[] = ['open', 'in_review', 'decided', 'closed'];

export const ENTITY_CONFIGS: readonly EntityConfig[] = [
  { key: 'transportContainer', path: 'containers', label: '运输容器', statuses: ['ready', 'in_transit', 'quarantine', 'cleared'] as const },
  { key: 'temperatureWindow', path: 'windows', label: '温控规则', statuses: ['draft', 'active', 'expired', 'superseded'] as const },
  { key: 'excursionEvent', path: 'excursions', label: '偏差事件', statuses: ['open', 'in_review', 'decided', 'closed'] as const },
  { key: 'dispositionDecision', path: 'dispositions', label: '处置决定', statuses: ['draft', 'release', 'quarantine', 'discard'] as const }
];
