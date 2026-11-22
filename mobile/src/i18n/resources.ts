import enCommon from './locales/en/common.json';
import esCommon from './locales/es/common.json';

export const defaultNS = 'common';

export const supportedLanguages = ['en', 'es'] as const;
export type Language = (typeof supportedLanguages)[number];

export function isLanguage(value: unknown): value is Language {
  return supportedLanguages.includes(value as Language);
}

/**
 * One JSON file per namespace and language. English is the source of truth:
 * the `satisfies` check fails type-checking if Spanish is missing a key.
 */
export const resources = {
  en: { common: enCommon },
  es: { common: esCommon satisfies typeof enCommon },
} as const;
