import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

import { defaultNS, resources, supportedLanguages } from './resources';

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    defaultNS,
    ns: [defaultNS],
    fallbackLng: 'en',
    supportedLngs: supportedLanguages,
    load: 'languageOnly',
    // React already escapes output.
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'area.lang',
      caches: ['localStorage'],
    },
  });

// Keep <html lang> in sync so screen readers use the right voice.
const syncHtmlLang = (lng: string) => {
  document.documentElement.lang = lng;
};
syncHtmlLang(i18n.resolvedLanguage ?? 'en');
i18n.on('languageChanged', syncHtmlLang);

export default i18n;
