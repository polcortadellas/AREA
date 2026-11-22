export type ConnectionResult =
  | { ok: true }
  | { ok: false; reason: 'timeout' | 'network' | 'notArea' }
  | { ok: false; reason: 'http'; status: number };

export const CONNECTION_TIMEOUT_MS = 5000;

/**
 * Checks that `baseUrl` is reachable and is an AREA server, using the API's
 * `GET /ping` health route (it answers "pong").
 */
export async function checkConnection(
  baseUrl: string,
  { timeoutMs = CONNECTION_TIMEOUT_MS, fetchImpl = fetch } = {},
): Promise<ConnectionResult> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetchImpl(`${baseUrl}/ping`, {
      method: 'GET',
      signal: controller.signal,
      headers: { Accept: 'text/plain' },
    });
    if (!res.ok) {
      return { ok: false, reason: 'http', status: res.status };
    }
    const body = await res.text();
    return body.trim() === 'pong'
      ? { ok: true }
      : { ok: false, reason: 'notArea' };
  } catch {
    return {
      ok: false,
      reason: controller.signal.aborted ? 'timeout' : 'network',
    };
  } finally {
    clearTimeout(timer);
  }
}
