import { QueryClient } from '@tanstack/react-query';

/** One factory so the app and tests build clients with the same defaults. */
export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  });
}
