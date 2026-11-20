import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { findBestLanguageTag } from 'react-native-localize';

import { readItem, storageKeys, writeItem } from '../storage/storage';
import {
  defaultNS,
  isLanguage,
  resources,
  supportedLanguages,
  type Language,
} from './resources';

function deviceLanguage(): Language {
  const best = findBestLanguageTag(supportedLanguages);
  return best && isLanguage(best.languageTag) ? best.languageTag : 'en';
}

// Initialised synchronously with the device language so the first frame is
// already translated; a language the user picked earlier is applied right after.
i18n.use(initReactI18next).init({
  resources,
  defaultNS,
  ns: [defaultNS],
  lng: deviceLanguage(),
  fallbackLng: 'en',
  supportedLngs: supportedLanguages,
  // React already escapes output.
  interpolation: { escapeValue: false },
});

/** Applies the language saved by the user, if any. */
export async function restoreLanguage(): Promise<void> {
  const saved = await readItem(storageKeys.language);
  if (isLanguage(saved) && saved !== i18n.language) {
    await i18n.changeLanguage(saved);
  }
}

export async function setLanguage(lng: Language): Promise<void> {
  await i18n.changeLanguage(lng);
  await writeItem(storageKeys.language, lng);
}

export default i18n;
