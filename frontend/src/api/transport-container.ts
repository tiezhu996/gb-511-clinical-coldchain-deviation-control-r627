
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listTransportContainer(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/containers?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createTransportContainer(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/containers', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionTransportContainer(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/containers/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
