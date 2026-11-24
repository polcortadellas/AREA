import { http, HttpResponse, type PathParams } from 'msw';

import type {
  Credentials,
  TokenResponse,
  User,
  VerifyEmailRequest,
} from '@/shared/api/generated/model';

/**
 * Behave like the real /auth endpoints (M-08), so the flows work in `npm run dev`
 * and tests: any e-mail logs in with `mockPassword`, `mockUnverifiedEmail` is
 * not confirmed yet, and only `mockVerifyToken` confirms an e-mail. The session
 * cookie is simulated with a flag kept for the browser tab.
 */
export const mockUser: User = {
  id: '4f8a2c1e-7b3d-4e9a-9c51-2d6f0b8e3a17',
  email: 'demo@example.com',
};
export const mockPassword = 'correct-horse';
export const mockUnverifiedEmail = 'unverified@example.com';
export const mockVerifyToken = 'valid-token';

const sessionKey = 'area.mock.session';

export function setMockSession(signedIn: boolean) {
  try {
    if (signedIn) window.sessionStorage.setItem(sessionKey, '1');
    else window.sessionStorage.removeItem(sessionKey);
  } catch {
    // Storage unavailable: the mock session simply does not persist.
  }
}

function hasMockSession() {
  try {
    return window.sessionStorage.getItem(sessionKey) === '1';
  } catch {
    return false;
  }
}

const token: TokenResponse = {
  access_token: 'mock-access-token',
  token_type: 'Bearer',
  expires_in: 3600,
};

const error = (status: number, code: string) =>
  HttpResponse.json({ error: code }, { status });

const signIn = () => {
  setMockSession(true);
  return HttpResponse.json(token);
};

export const authHandlers = [
  http.post('*/auth/register', () =>
    HttpResponse.json({ status: 'verification_email_sent' }, { status: 202 }),
  ),
  http.post<PathParams, VerifyEmailRequest>(
    '*/auth/verify-email',
    async ({ request }) => {
      const body = await request.json();
      return body.token === mockVerifyToken
        ? signIn()
        : error(400, 'invalid_token');
    },
  ),
  http.post<PathParams, Credentials>('*/auth/login', async ({ request }) => {
    const body = await request.json();
    if (body.password !== mockPassword) {
      return error(401, 'invalid_credentials');
    }
    if (body.email === mockUnverifiedEmail) {
      return error(403, 'email_not_verified');
    }
    return signIn();
  }),
  http.post('*/auth/logout', () => {
    setMockSession(false);
    return new HttpResponse(null, { status: 204 });
  }),
  http.get('*/auth/me', () =>
    hasMockSession() ? HttpResponse.json(mockUser) : error(401, 'unauthorized'),
  ),
];
