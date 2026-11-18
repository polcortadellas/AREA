import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll, beforeEach } from 'vitest';

import { server } from '@/mocks/node';
import { i18n } from '@/shared/i18n';

// Same MSW handlers as `npm run dev`; any request without a handler fails the test.
beforeAll(() => server.listen({ onUnhandledFrame: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

beforeEach(async () => {
  window.localStorage.clear();
  await i18n.changeLanguage('en');
});
afterEach(() => cleanup());
