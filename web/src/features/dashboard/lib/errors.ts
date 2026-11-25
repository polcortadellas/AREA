import { ApiError } from '@/shared/api/http';

/** Maps a dashboard API error to an i18n message key. */
export function dashboardErrorKey(error: unknown) {
  if (!(error instanceof ApiError)) {
    return error instanceof TypeError
      ? ('dashboard.errors.network' as const)
      : ('dashboard.errors.unknown' as const);
  }
  if (error.status === 401) return 'dashboard.errors.unauthorized' as const;
  if (error.status === 404) return 'dashboard.errors.notFound' as const;
  return 'dashboard.errors.unknown' as const;
}
