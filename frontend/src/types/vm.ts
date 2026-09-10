export type VMFamily = 'sm' | 'cm' | 'ws' | 'oa' | 'unknown';

// Ces types reflètent le JSON renvoyé par le backend (voir internal/store/store.go).
export interface VM {
  hostname: string;
  ip: string;
  family: string;
  hypervisor: string;
  app_versions?: Record<string, string>;
  last_seen: string;
  last_error?: string;
}

export interface Group {
  id: string;
  vms: Record<string, VM>; // clé = family
  in_use_by?: string;
  checked_out_at?: string;
}
