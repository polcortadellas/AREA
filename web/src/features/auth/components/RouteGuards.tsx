import { useTranslation } from 'react-i18next';
import { Navigate, Outlet, useLocation, useSearchParams } from 'react-router';

import { useCurrentUser } from '../hooks/useCurrentUser';
import { safeNext } from '../lib/next';

function Loading() {
  const { t } = useTranslation();
  return (
    <p role="status" className="text-muted-foreground">
      {t('layout.loading')}
    </p>
  );
}

/** Wraps routes that need a logged-in user; others are sent to /login and brought back after. */
export function RequireAuth() {
  const { data: user, isPending } = useCurrentUser();
  const { pathname, search } = useLocation();

  if (isPending) return <Loading />;
  if (!user) {
    const next = pathname + search;
    const to =
      next === '/' ? '/login' : `/login?next=${encodeURIComponent(next)}`;
    return <Navigate to={to} replace />;
  }
  return <Outlet />;
}

/** Wraps /login and /register: a logged-in user goes straight to where they were heading. */
export function GuestOnly() {
  const { data: user, isPending } = useCurrentUser();
  const [searchParams] = useSearchParams();

  if (isPending) return <Loading />;
  if (user) {
    return <Navigate to={safeNext(searchParams.get('next'))} replace />;
  }
  return <Outlet />;
}
