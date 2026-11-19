import { Link } from 'react-router';
import { useTranslation } from 'react-i18next';

import { Button } from '@/shared/ui/button';

export function NotFoundPage() {
  const { t } = useTranslation();

  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-semibold">{t('notFound.title')}</h1>
      <p className="text-muted-foreground">{t('notFound.description')}</p>
      <Button asChild variant="outline">
        <Link to="/">{t('notFound.back')}</Link>
      </Button>
    </section>
  );
}
