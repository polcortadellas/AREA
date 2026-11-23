import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';

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

  it('sends the session cookie and the CSRF header', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch');

    await apiFetch('/about.json');

    expect(fetchSpy.mock.calls[0]?.[1]).toMatchObject({
      credentials: 'include',
      headers: { 'X-Requested-With': 'XMLHttpRequest' },
    });
  });

  it('still throws an ApiError when the error body is not JSON', async () => {
    const html = '<html><body>502 Bad Gateway</body></html>';
    server.use(
      http.get('*/about.json', () => HttpResponse.html(html, { status: 502 })),
    );

    const error = await apiFetch('/about.json').catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 502, body: html });
  });
});
