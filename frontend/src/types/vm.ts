export type VMFamily = 'sm' | 'cm' | 'ws' | 'oa' | 'unknown';

export interface VM {
  id: string;
  ip: string;
  hostname: string;
  family: VMFamily;
  hypervisor: string;
  groupId: string;
  appVersions: Record<string, string>;
  lastChecked: string;
  reachable: boolean;
  error?: string;
}

export interface Group {
  id: string;
  vms: VM[];
  inUseBy?: string;
  checkedAt?: string;
}
