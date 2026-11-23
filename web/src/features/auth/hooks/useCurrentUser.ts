import { queryOptions, useQuery } from '@tanstack/react-query';

import { getGetMeQueryKey, getMe } from '@/shared/api/generated/auth/auth';
import { ApiError } from '@/shared/api/http';

/**
 * The logged-in user, or null when the session cookie is missing or expired.
 * The cookie is HttpOnly, so asking /auth/me is the only way to know.
 */
export function currentUserQueryOptions() {
  return queryOptions({
    queryKey: getGetMeQueryKey(),
    queryFn: async ({ signal }) => {
      try {
        return await getMe({ signal });
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) return null;
        throw error;
      }
    },
  });
}

export function useCurrentUser() {
  return useQuery(currentUserQueryOptions());
}
