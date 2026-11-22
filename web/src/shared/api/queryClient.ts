import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query';

import { getGetMeQueryKey } from './generated/auth/auth';
import { ApiError } from './http';

/** One factory so the app and tests build clients with the same defaults. */
export function createQueryClient() {
  // Any 401 means the session cookie expired or was revoked: forgetting the
  // current user sends protected pages back to the login page.
  const onError = (error: unknown) => {
    if (error instanceof ApiError && error.status === 401) {
      client.setQueryData(getGetMeQueryKey(), null);
    }
  };

  const client: QueryClient = new QueryClient({
    queryCache: new QueryCache({ onError }),
    mutationCache: new MutationCache({ onError }),
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  });
  return client;
}
