import { useCallback } from 'react';
import { useQueryClient, type Query } from '@tanstack/react-query';
import { useNavigate } from 'react-router';

import { getGetMeQueryKey, useLogout } from '@/shared/api/generated/auth/auth';
import { currentUserQueryOptions } from './useCurrentUser';

// Cached data is dropped on every login and logout so one user never sees another's data.

const isCurrentUser = (query: Query) =>
  query.queryKey[0] === getGetMeQueryKey()[0];

/** Call after the server set the session cookie (login, email confirmation). */
export function useStartSession() {
  const queryClient = useQueryClient();
  return useCallback(async () => {
    await queryClient.resetQueries();
    await queryClient.fetchQuery(currentUserQueryOptions());
  }, [queryClient]);
}

export function useEndSession() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const logout = useLogout();
  return useCallback(() => {
    logout.mutate(undefined, {
      onSuccess: () => {
        queryClient.removeQueries({ predicate: q => !isCurrentUser(q) });
        queryClient.setQueryData(getGetMeQueryKey(), null);
        void navigate('/login', { replace: true });
      },
    });
  }, [logout, queryClient, navigate]);
}
