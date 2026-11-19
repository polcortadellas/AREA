import 'i18next';

import type { defaultNS, resources } from './resources';

// Makes t('...') keys type-checked against the English JSON files.
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: typeof defaultNS;
    resources: (typeof resources)['en'];
  }
}
