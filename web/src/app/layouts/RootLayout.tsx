import { NavLink, Outlet } from 'react-router';
import { useTranslation } from 'react-i18next';

import { LanguageSwitcher } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';

export function RootLayout() {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-svh flex-col">
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:rounded-md focus:bg-background focus:px-3 focus:py-2 focus:ring-[3px] focus:ring-ring/50"
      >
        {t('layout.skipToContent')}
      </a>

      <header className="border-b">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
          <NavLink to="/" className="text-lg font-semibold">
            {t('appName')}
          </NavLink>
          <nav aria-label={t('layout.mainNav')}>
            <ul className="flex gap-4 text-sm">
              <li>
                <NavLink
                  to="/"
                  end
                  className={({ isActive }) =>
                    cn('hover:underline', isActive && 'font-semibold underline')
                  }
                >
                  {t('nav.dashboard')}
                </NavLink>
              </li>
            </ul>
          </nav>
          <LanguageSwitcher />
        </div>
      </header>

      <main
        id="main"
        tabIndex={-1}
        className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 outline-none"
      >
        <Outlet />
      </main>
    </div>
  );
}
