import { useTranslation } from 'react-i18next';

import { Button } from '@/shared/ui/button';

/** Shown when a route throws while rendering or loading. */
export function RouteErrorPage() {
  const { t } = useTranslation();

  return (
    <main role="alert" className="mx-auto max-w-xl space-y-4 p-6">
      <h1 className="text-2xl font-semibold">{t('error.title')}</h1>
      <p className="text-muted-foreground">{t('error.description')}</p>
      <Button onClick={() => window.location.reload()}>
        {t('error.reload')}
      </Button>
    </main>
  );
}
