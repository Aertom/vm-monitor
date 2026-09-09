import { useQuery } from '@tanstack/react-query';
import type { Group } from '../types/vm';

export function useGroups() {
  return useQuery<Group[]>({
    queryKey: ['groups'],
    queryFn: async () => {
      const res = await fetch('/api/groups');
      if (!res.ok) throw new Error('failed to fetch groups');
      return res.json();
    },
    refetchInterval: 15000,
  });
}
