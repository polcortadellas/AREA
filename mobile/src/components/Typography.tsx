import { StyleSheet, Text, type TextProps } from 'react-native';

import { fontSize, useTheme } from '../theme/theme';

/** Section title. TalkBack lists headers so users can jump between them. */
export function Heading({ style, ...props }: TextProps) {
  const { colors } = useTheme();
  return (
    <Text
      accessibilityRole="header"
      style={[styles.heading, { color: colors.text }, style]}
      {...props}
    />
  );
}

export function Body({
  muted = false,
  style,
  ...props
}: TextProps & { muted?: boolean }) {
  const { colors } = useTheme();
  return (
    <Text
      style={[
        styles.body,
        { color: muted ? colors.textMuted : colors.text },
        style,
      ]}
      {...props}
    />
  );
}

const styles = StyleSheet.create({
  heading: { fontSize: fontSize.heading, fontWeight: '700' },
  body: { fontSize: fontSize.body, lineHeight: 22 },
});
