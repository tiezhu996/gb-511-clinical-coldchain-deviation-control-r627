import { create } from 'zustand';
import { listSensorEvidence } from '../api/sensor-evidence';
import type { SensorEvidence } from '../types/domain';

export const useSensorEvidenceStore = create<{ items: SensorEvidence[]; error: string; load: () => Promise<void> }>((set) => ({
  items: [], error: '',
  load: async () => {
    try { set({ items: (await listSensorEvidence()).data, error: '' }); }
    catch (error) { set({ error: error instanceof Error ? error.message : String(error) }); }
  },
}));
