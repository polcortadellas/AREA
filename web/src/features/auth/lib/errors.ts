import { ApiError } from '@/shared/api/http';

/** Maps an auth request failure to the i18n key of the message to show. */
export function authErrorKey(error: unknown) {
  if (!(error instanceof ApiError)) {
    return error instanceof TypeError
      ? ('auth.errors.network' as const)
      : ('auth.errors.unknown' as const);
  }
  const code = (error.body as { error?: unknown } | undefined)?.error;
  switch (code) {
    case 'invalid_credentials':
      return 'auth.errors.invalidCredentials' as const;
    case 'email_not_verified':
      return 'auth.errors.emailNotVerified' as const;
    case 'invalid_token':
      return 'auth.errors.invalidToken' as const;
    case 'invalid_email':
      return 'auth.errors.invalidEmail' as const;
    case 'weak_password':
      return 'auth.errors.weakPassword' as const;
    default:
      return 'auth.errors.unknown' as const;
  }
}
