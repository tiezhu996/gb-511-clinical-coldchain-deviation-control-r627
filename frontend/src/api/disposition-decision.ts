
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listDispositionDecision(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/dispositions?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createDispositionDecision(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/dispositions', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionDispositionDecision(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/dispositions/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
