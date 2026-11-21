import { StyleSheet, Text, View } from 'react-native';

import { fontSize, spacing, useTheme } from '../theme/theme';

type Props = {
  tone: 'info' | 'success' | 'error';
  message: string;
  testID?: string;
};

const icons = { info: 'ℹ', success: '✓', error: '✕' } as const;

/**
 * Inline status line announced by TalkBack when it changes. The icon is
 * decorative, so meaning never relies on colour alone.
 */
export function StatusMessage({ tone, message, testID }: Props) {
  const { colors } = useTheme();
  const color =
    tone === 'error'
      ? colors.danger
      : tone === 'success'
        ? colors.success
        : colors.textMuted;

  return (
    <View
      testID={testID}
      accessible
      accessibilityLiveRegion={tone === 'error' ? 'assertive' : 'polite'}
      accessibilityRole={tone === 'error' ? 'alert' : 'text'}
      accessibilityLabel={message}
      style={[
        styles.row,
        { backgroundColor: colors.surface, borderLeftColor: color },
      ]}
    >
      <Text importantForAccessibility="no" style={[styles.icon, { color }]}>
        {icons[tone]}
      </Text>
      <Text style={[styles.text, { color: colors.text }]}>{message}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    gap: spacing.sm,
    padding: spacing.md,
    borderRadius: 8,
    borderLeftWidth: 4,
  },
  icon: { fontSize: fontSize.body, fontWeight: '700' },
  text: { flex: 1, fontSize: fontSize.body },
});
