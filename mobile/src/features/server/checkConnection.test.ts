import { checkConnection } from './checkConnection';

const respond = (status: number, body: string) =>
  jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(body),
  });

describe('checkConnection', () => {
  it('calls GET /ping on the base URL and accepts "pong"', async () => {
    const fetchImpl = respond(200, 'pong\n');
    await expect(
      checkConnection('http://10.0.2.2:8080', { fetchImpl }),
    ).resolves.toEqual({
      ok: true,
    });
    expect(fetchImpl).toHaveBeenCalledWith(
      'http://10.0.2.2:8080/ping',
      expect.objectContaining({ method: 'GET' }),
    );
  });

  it('reports a server that is not AREA', async () => {
    await expect(
      checkConnection('http://x', {
        fetchImpl: respond(200, '<html>router</html>'),
      }),
    ).resolves.toEqual({ ok: false, reason: 'notArea' });
  });

  it('reports HTTP errors with their status', async () => {
    await expect(
      checkConnection('http://x', { fetchImpl: respond(502, 'Bad Gateway') }),
    ).resolves.toEqual({
      ok: false,
      reason: 'http',
      status: 502,
    });
  });

  it('reports network errors', async () => {
    const fetchImpl = jest
      .fn()
      .mockRejectedValue(new TypeError('Network request failed'));
    await expect(checkConnection('http://x', { fetchImpl })).resolves.toEqual({
      ok: false,
      reason: 'network',
    });
  });

  it('gives up after the timeout', async () => {
    const fetchImpl = jest.fn(
      (_url: string, init: RequestInit) =>
        new Promise((_resolve, reject) => {
          init.signal?.addEventListener('abort', () =>
            reject(new Error('Aborted')),
          );
        }),
    );
    await expect(
      checkConnection('http://x', {
        fetchImpl: fetchImpl as unknown as typeof fetch,
        timeoutMs: 10,
      }),
    ).resolves.toEqual({ ok: false, reason: 'timeout' });
  });
});
