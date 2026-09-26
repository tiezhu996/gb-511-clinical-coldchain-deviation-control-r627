
import { request } from './client';
import type { AuditLog } from '../types/domain';
export async function listAudits(page = 1, pageSize = 30) {
  return request<AuditLog[]>(`/audits?page=${page}&pageSize=${pageSize}`);
}
export async function loadOverview() { return request<Record<string, Record<string, number>>>('/overview'); }
