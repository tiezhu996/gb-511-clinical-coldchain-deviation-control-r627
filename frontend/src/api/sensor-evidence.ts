import { request } from './client';
import type { SensorEvidence } from '../types/domain';

export async function listSensorEvidence() {
  return request<SensorEvidence[]>('/evidence?page=1&pageSize=100');
}

export async function registerSensorEvidence(input: Partial<SensorEvidence>) {
  return request<SensorEvidence>('/evidence', { method: 'POST', body: JSON.stringify(input) });
}
