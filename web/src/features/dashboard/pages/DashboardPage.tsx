import { useTranslation } from 'react-i18next';

// Placeholder until M-21 (workflow list, toggle, history).
export function DashboardPage() {
  const { t } = useTranslation();

  return (
    <section>
      <h1 className="text-2xl font-semibold">{t('dashboard.title')}</h1>
    </section>
  );
}
