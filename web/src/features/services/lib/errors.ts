import { ApiError } from '@/shared/api/http';

/** Maps a failed link or unlink request to the i18n key of the message to show. */
export function serviceErrorKey(error: unknown, action: 'link' | 'unlink') {
  if (!(error instanceof ApiError)) {
    return error instanceof TypeError
      ? ('services.errors.network' as const)
      : ('services.errors.unknown' as const);
  }
  if (error.status === 404) {
    // Link: no OAuth app configured on the server. Unlink: already unlinked.
    return action === 'link'
      ? ('services.errors.notConfigured' as const)
      : ('services.errors.notLinked' as const);
  }
  return 'services.errors.unknown' as const;
}

/** Message for the `?error=` code the OAuth callback redirects with. */
export function callbackErrorKey(code: string) {
  switch (code) {
    case 'access_denied':
      return 'services.callback.accessDenied' as const;
    case 'invalid_state':
      return 'services.callback.invalidState' as const;
    default:
      return 'services.callback.linkFailed' as const;
  }
}
