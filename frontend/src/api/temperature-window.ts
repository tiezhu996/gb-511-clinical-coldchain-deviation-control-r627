
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listTemperatureWindow(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/windows?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createTemperatureWindow(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/windows', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionTemperatureWindow(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/windows/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
