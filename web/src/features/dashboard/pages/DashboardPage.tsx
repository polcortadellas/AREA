import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';

import { Button } from '@/shared/ui/button';

// Placeholder until M-21 (workflow list, toggle, history).
export function DashboardPage() {
  const { t } = useTranslation();

  return (
    <section className="flex flex-wrap items-center justify-between gap-4">
      <h1 className="text-2xl font-semibold">{t('dashboard.title')}</h1>
      <Button asChild>
        <Link to="/workflows/new">{t('dashboard.newWorkflow')}</Link>
      </Button>
    </section>
  );
}
