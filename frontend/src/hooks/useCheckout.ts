import { useMutation, useQueryClient } from '@tanstack/react-query';

interface CheckoutParams {
  groupId: string;
  user: string;
}

export function useCheckout() {
  const queryClient = useQueryClient();

  const checkoutMutation = useMutation({
    mutationFn: async ({ groupId, user }: CheckoutParams) => {
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

  const checkinMutation = useMutation({
    mutationFn: async (groupId: string) => {
      const res = await fetch(`/api/groups/${groupId}/checkin`, {
        method: 'POST',
      });
      if (!res.ok) throw new Error(await res.text());
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['groups'] }),
  });

  const checkout = (groupId: string, user: string) => checkoutMutation.mutate({ groupId, user });
  const checkin = (groupId: string) => checkinMutation.mutate(groupId);

  return {
    checkout,
    checkin,
    isPending: checkoutMutation.isPending || checkinMutation.isPending,
  };
}
