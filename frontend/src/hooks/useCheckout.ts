import { useMutation, useQueryClient } from '@tanstack/react-query';

export function useCheckout(groupId: string) {
  const queryClient = useQueryClient();

  const checkout = useMutation({
    mutationFn: async (user: string) => {
      const res = await fetch(`/api/groups/${groupId}/checkout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user }),
      });
      if (!res.ok) {
        throw new Error(await res.text());
      }
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['groups'] }),
  });

  const checkin = useMutation({
    mutationFn: async () => {
      const res = await fetch(`/api/groups/${groupId}/checkin`, {
        method: 'POST',
      });
      if (!res.ok) throw new Error(await res.text());
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['groups'] }),
  });

  return { checkout: checkout.mutate, checkin: checkin.mutate };
}
