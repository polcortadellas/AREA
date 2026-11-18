import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { server } from '@/mocks/node';
import { ApiError, apiFetch } from './http';

describe('apiFetch', () => {
  it('returns the parsed JSON body on success', async () => {
    const about = await apiFetch<{ server: { services: unknown[] } }>(
      '/about.json',
    );

    expect(about.server.services.length).toBeGreaterThan(0);
  });

  it('throws an ApiError carrying the status and error body', async () => {
    const details = { errors: [{ path: 'steps[0]', message: 'required' }] };
    server.use(
      http.post('*/workflows', () =>
        HttpResponse.json(details, { status: 422 }),
      ),
    );

    const error = await apiFetch('/workflows', { method: 'POST' }).catch(
      (e: unknown) => e,
    );

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 422, body: details });
  });
});
