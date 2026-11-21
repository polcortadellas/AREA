import { useId } from 'react';
import {
  StyleSheet,
  Text,
  TextInput,
  View,
  type TextInputProps,
} from 'react-native';

import { fontSize, minTouchTarget, spacing, useTheme } from '../theme/theme';

type Props = Omit<TextInputProps, 'style' | 'accessibilityLabel'> & {
  label: string;
  hint?: string;
  error?: string | null;
};

/**
 * Labelled text input. The visible label is also the accessible name, the hint
 * is read after it, and an error is announced as soon as it appears.
 */
export function TextField({ label, hint, error, ...inputProps }: Props) {
  const { colors } = useTheme();
  const labelId = useId();

  return (
    <View style={styles.container}>
      <Text nativeID={labelId} style={[styles.label, { color: colors.text }]}>
        {label}
      </Text>
      {hint ? (
        <Text style={[styles.hint, { color: colors.textMuted }]}>{hint}</Text>
      ) : null}
      <TextInput
        {...inputProps}
        accessibilityLabel={label}
        accessibilityLabelledBy={labelId}
        accessibilityHint={error ?? hint}
        accessibilityState={{ disabled: inputProps.editable === false }}
        placeholderTextColor={colors.textMuted}
        style={[
          styles.input,
          {
            color: colors.text,
            backgroundColor: colors.background,
            borderColor: error ? colors.danger : colors.border,
          },
        ]}
      />
      {error ? (
        <Text
          accessibilityLiveRegion="polite"
          accessibilityRole="alert"
          style={[styles.error, { color: colors.danger }]}
        >
          {error}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: spacing.xs },
  label: { fontSize: fontSize.body, fontWeight: '600' },
  hint: { fontSize: fontSize.small },
  input: {
    minHeight: minTouchTarget,
    borderWidth: 2,
    borderRadius: 8,
    paddingHorizontal: spacing.md,
    fontSize: fontSize.body,
  },
  error: { fontSize: fontSize.small, fontWeight: '600' },
});
