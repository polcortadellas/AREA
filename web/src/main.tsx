import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from '@/app/App';
import { enableMocking } from '@/mocks/enableMocking';
import '@/shared/i18n/config';
import './index.css';

if (import.meta.env.DEV) {
  // Logs accessibility violations to the console; never part of production builds.
  void import('@/shared/a11y/devAxe').then(({ startDevAxe }) => startDevAxe());
}

void enableMocking().then(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
});
