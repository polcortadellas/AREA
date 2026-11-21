import { useId } from 'react';
import { useTranslation } from 'react-i18next';

import { cn } from '@/shared/lib/utils';
import { supportedLanguages, type Language } from './resources';

export function LanguageSwitcher({ className }: { className?: string }) {
  const { t, i18n } = useTranslation();
  const id = useId();
  const current = (i18n.resolvedLanguage ?? 'en') as Language;

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <label htmlFor={id} className="text-sm">
        {t('language.label')}
      </label>
      <select
        id={id}
        value={current}
        onChange={e => void i18n.changeLanguage(e.target.value)}
        className="h-9 rounded-md border border-input bg-background px-2 text-sm focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-none"
      >
        {supportedLanguages.map(lng => (
          <option key={lng} value={lng} lang={lng}>
            {t(`language.${lng}`)}
          </option>
        ))}
      </select>
    </div>
  );
}
