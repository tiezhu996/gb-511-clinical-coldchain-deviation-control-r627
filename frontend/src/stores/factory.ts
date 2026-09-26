
import { create } from 'zustand';
import { request } from '../api/client';
import type { ApiEnvelope, DomainRecord, PageMeta } from '../types/domain';

export interface EntityState {
  items: DomainRecord[];
  meta: PageMeta;
  loading: boolean;
  error: string;
  load: (path: string, search?: string) => Promise<void>;
  createRecord: (path: string, input: Partial<DomainRecord>) => Promise<void>;
  transition: (path: string, item: DomainRecord, status: string, reason?: string, evidence?: string) => Promise<void>;
}
export type EntityStore = ReturnType<typeof createEntityStore>;

export function createEntityStore() {
  return create<EntityState>((set, get) => ({
    items: [], meta: { page: 1, pageSize: 20, total: 0 }, loading: false, error: '',
    load: async (path, search = '') => {
      set({ loading: true, error: '' });
      try {
        const result = await request<DomainRecord[]>(`/${path}?page=1&pageSize=20&search=${encodeURIComponent(search)}`);
        set({ items: result.data, meta: result.meta || { page: 1, pageSize: 20, total: result.data.length }, loading: false });
      } catch (error) { set({ error: error instanceof Error ? error.message : String(error), loading: false }); }
    },
    createRecord: async (path, input) => {
      set({ loading: true, error: '' });
      try {
        await request<DomainRecord>(`/${path}`, { method: 'POST', body: JSON.stringify(input) });
        await get().load(path);
      } catch (error) { set({ error: error instanceof Error ? error.message : String(error), loading: false }); throw error; }
    },
    transition: async (path, item, status, reason = '质量工作台人工确认', evidence = '') => {
      set({ loading: true, error: '' });
      try {
        await request<DomainRecord>(`/${path}/${item.id}/transition`, { method: 'POST', body: JSON.stringify({ status, expectedVersion: item.version, reason, evidence }) });
        await get().load(path);
      } catch (error) { set({ error: error instanceof Error ? error.message : String(error), loading: false }); throw error; }
    },
  }));
}
