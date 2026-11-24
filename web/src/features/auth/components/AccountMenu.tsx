import { Link } from 'react-router';
import { useTranslation } from 'react-i18next';

import { Button } from '@/shared/ui/button';
import { useCurrentUser } from '../hooks/useCurrentUser';
import { useEndSession } from '../hooks/useSessionActions';

export function AccountMenu() {
  const { t } = useTranslation();
  const { data: user, isPending } = useCurrentUser();
  const endSession = useEndSession();

  if (isPending) return null;

  if (!user) {
    return (
      <div className="flex items-center gap-2">
        <Button asChild variant="ghost" size="sm">
          <Link to="/login">{t('nav.login')}</Link>
        </Button>
        <Button asChild size="sm">
          <Link to="/register">{t('nav.register')}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3">
      <span className="text-muted-foreground hidden text-sm sm:inline">
        {t('auth.signedInAs', { email: user.email })}
      </span>
      <Button variant="outline" size="sm" onClick={endSession}>
        {t('nav.logout')}
      </Button>
    </div>
  );
}
