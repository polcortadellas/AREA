export const apiUrl = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '');

/** Error thrown for any non-2xx response; `body` holds the parsed JSON error (e.g. HTTP 422 details), or the raw text if it isn't JSON. */
export class ApiError<TBody = unknown> extends Error {
  readonly status: number;
  readonly body: TBody;

  constructor(status: number, body: TBody) {
    super(`API request failed with status ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

/**
 * The single place that talks to the server. Orval-generated hooks call it,
 * so auth headers (M-18) and error handling only need to be added here.
 */
export async function apiFetch<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiUrl}${url}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...init?.headers,
    },
  });

  const text = await response.text();

  if (!response.ok) {
    throw new ApiError(response.status, parseErrorBody(text));
  }
  return (text ? JSON.parse(text) : undefined) as T;
}

/** Error bodies aren't always JSON (e.g. a proxy's 502 HTML page): fall back to the raw text. */
function parseErrorBody(text: string): unknown {
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// Orval picks these up so generated hooks type `error` as ApiError.
export type ErrorType<TBody> = ApiError<TBody>;
export type BodyType<TBody> = TBody;
