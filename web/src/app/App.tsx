import { useTranslation } from 'react-i18next';

import { LanguageSwitcher } from '@/shared/i18n';

export function App() {
  const { t } = useTranslation();

  return (
    <main className="p-6">
      <h1 className="mb-4 text-2xl font-semibold">{t('appName')}</h1>
      <LanguageSwitcher />
    </main>
  );
}
