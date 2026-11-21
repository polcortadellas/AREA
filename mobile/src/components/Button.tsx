import { ActivityIndicator, Pressable, StyleSheet, Text } from 'react-native';

import { fontSize, minTouchTarget, spacing, useTheme } from '../theme/theme';

type Props = {
  label: string;
  onPress: () => void;
  variant?: 'primary' | 'secondary';
  disabled?: boolean;
  /** Shows a spinner and tells screen readers the action is in progress. */
  busy?: boolean;
  accessibilityHint?: string;
  testID?: string;
};

export function Button({
  label,
  onPress,
  variant = 'primary',
  disabled = false,
  busy = false,
  accessibilityHint,
  testID,
}: Props) {
  const { colors } = useTheme();
  const primary = variant === 'primary';
  const inactive = disabled || busy;
  const fg = primary ? colors.onPrimary : colors.primary;

  return (
    <Pressable
      testID={testID}
      onPress={onPress}
      disabled={inactive}
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityHint={accessibilityHint}
      accessibilityState={{ disabled: inactive, busy }}
      android_ripple={{ color: primary ? colors.onPrimary : colors.primary }}
      style={({ pressed }) => [
        styles.base,
        primary
          ? { backgroundColor: colors.primary }
          : { borderColor: colors.primary, borderWidth: 2 },
        (pressed || inactive) && styles.dimmed,
      ]}
    >
      {busy ? <ActivityIndicator color={fg} /> : null}
      <Text style={[styles.label, { color: fg }]}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    minHeight: minTouchTarget,
    paddingHorizontal: spacing.md,
    borderRadius: 8,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
  },
  label: { fontSize: fontSize.body, fontWeight: '600' },
  dimmed: { opacity: 0.7 },
});
