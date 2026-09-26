
export interface DomainRecord {
  id: number;
  code: string;
  name: string;
  status: string;
  version: number;
  description: string;
  facility: string;
  owner: string;
  category: string;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  metricValue: number;
  metricUnit: string;
  effectiveAt: string;
  evidence: string;
  relatedCode: string;
  sensorId?: string;
  containerType?: string;
  currentLocation?: string;
  custodian?: string;
  currentTempC?: number;
  lastSensorReading?: string;
  productClass?: string;
  minimumCelsius?: number;
  maximumCelsius?: number;
  maxExcursionMinutes?: number;
  qualityOwner?: string;
  containerCode?: string;
  windowCode?: string;
  observedTempC?: number;
  durationMinutes?: number;
  detectedAt?: string;
  sensorEvidence?: string;
  reviewer?: string;
  excursionCode?: string;
  decisionBasis?: string;
  proposedBy?: string;
  approvedBy?: string;
  decidedAt?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PageMeta { page: number; pageSize: number; total: number }
export interface ApiEnvelope<T> { data: T; error?: string; message?: string; meta?: PageMeta }
export interface UserSession { token: string; username: string; displayName: string; role: string; expiresIn: number }
export interface SessionProfile { username: string; displayName: string; role: string; requestId: string }
export interface SensorEvidence {
  id: number; code: string; excursionCode: string; containerCode: string; objectKey: string;
  sha256: string; mediaType: string; sizeBytes: number; capturedAt: string; capturedBy: string; source: string; uploadUrl?: string;
}
export interface AuditLog {
  id: number; requestId: string; actor: string; action: string; entityType: string;
  entityId: number; beforeState: string; afterState: string; detail: string; createdAt: string;
}
export interface EntityConfig { key: string; path: string; label: string; statuses: readonly string[] }

export type Role = 'viewer' | 'operator' | 'reviewer' | 'admin';

export function roleAtLeast(role: string | undefined, minimum: Role): boolean {
  const rank: Record<Role, number> = { viewer: 1, operator: 2, reviewer: 3, admin: 4 };
  return Boolean(role && rank[role as Role] >= rank[minimum]);
}
