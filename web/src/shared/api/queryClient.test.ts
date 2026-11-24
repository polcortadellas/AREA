import { describe, expect, it } from 'vitest';

import { getGetMeQueryKey } from './generated/auth/auth';
import { ApiError } from './http';
import { createQueryClient } from './queryClient';

describe('createQueryClient', () => {
  it('forgets the current user when any request answers 401', async () => {
    const client = createQueryClient();
    client.setQueryData(getGetMeQueryKey(), { id: '1', email: 'a@b.c' });

    await client
      .fetchQuery({
        queryKey: ['workflows'],
        queryFn: () =>
          Promise.reject(new ApiError(401, { error: 'unauthorized' })),
        retry: false,
      })
      .catch(() => undefined);

    expect(client.getQueryData(getGetMeQueryKey())).toBeNull();
  });

  it('keeps the user on other errors', async () => {
    const client = createQueryClient();
    const user = { id: '1', email: 'a@b.c' };
    client.setQueryData(getGetMeQueryKey(), user);

    await client
      .fetchQuery({
        queryKey: ['workflows'],
        queryFn: () => Promise.reject(new ApiError(500, undefined)),
        retry: false,
      })
      .catch(() => undefined);

    expect(client.getQueryData(getGetMeQueryKey())).toEqual(user);
  });
});
