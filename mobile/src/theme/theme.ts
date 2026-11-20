import { useColorScheme } from 'react-native';

/**
 * Colour tokens for light and dark mode. Every text/background pair used by the
 * components meets WCAG AA (4.5:1 for body text, 3:1 for borders and icons):
 *
 * light  text #111827 on #FFFFFF 17.7:1 · muted #4B5563 on #FFFFFF 7.6:1
 *        primary #1D4ED8 on #FFFFFF 6.7:1 · onPrimary #FFFFFF on #1D4ED8 6.7:1
 *        danger #B91C1C on #FFFFFF 6.5:1  · success #15803D on #FFFFFF 5.0:1
 * dark   text #F9FAFB on #0B1220 17.9:1 · muted #CBD5E1 on #0B1220 12.6:1
 *        primary #93C5FD on #0B1220 10.4:1 · onPrimary #0B1220 on #93C5FD 10.4:1
 *        danger #FCA5A5 on #0B1220 9.9:1  · success #86EFAC on #0B1220 13.3:1
 */
const light = {
  background: '#FFFFFF',
  surface: '#F3F4F6',
  text: '#111827',
  textMuted: '#4B5563',
  border: '#6B7280',
  primary: '#1D4ED8',
  onPrimary: '#FFFFFF',
  danger: '#B91C1C',
  success: '#15803D',
} as const;

const dark: Record<keyof typeof light, string> = {
  background: '#0B1220',
  surface: '#1F2937',
  text: '#F9FAFB',
  textMuted: '#CBD5E1',
  border: '#94A3B8',
  primary: '#93C5FD',
  onPrimary: '#0B1220',
  danger: '#FCA5A5',
  success: '#86EFAC',
};

export type Colors = Record<keyof typeof light, string>;

export const spacing = { xs: 4, sm: 8, md: 16, lg: 24, xl: 32 } as const;

/** Android's minimum recommended touch target, in dp. */
export const minTouchTarget = 48;

export const fontSize = {
  body: 16,
  small: 14,
  title: 22,
  heading: 18,
} as const;

export function useTheme() {
  const scheme = useColorScheme();
  const colors: Colors = scheme === 'dark' ? dark : light;
  return { colors, dark: scheme === 'dark' };
}
