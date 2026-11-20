import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useTranslation } from 'react-i18next';

import { setLanguage, supportedLanguages, type Language } from '../i18n';
import { fontSize, minTouchTarget, spacing, useTheme } from '../theme/theme';

/** Radio group listing every supported language, each named in its own language. */
export function LanguageSwitcher() {
  const { t, i18n } = useTranslation();
  const { colors } = useTheme();
  const current = i18n.resolvedLanguage as Language;

  return (
    <View
      accessibilityRole="radiogroup"
      accessibilityLabel={t('language.label')}
      style={styles.group}
    >
      {supportedLanguages.map(lng => {
        const selected = lng === current;
        // Native name ("English", "Español") so users can find their language
        // even when the UI is in a language they do not read.
        const name = t(`language.${lng}`, { lng });
        return (
          <Pressable
            key={lng}
            testID={`language-${lng}`}
            onPress={() => setLanguage(lng)}
            accessibilityRole="radio"
            accessibilityLabel={name}
            accessibilityState={{ checked: selected }}
            accessibilityLanguage={lng}
            style={[
              styles.option,
              { borderColor: selected ? colors.primary : colors.border },
            ]}
          >
            <View
              importantForAccessibility="no"
              style={[
                styles.radio,
                { borderColor: selected ? colors.primary : colors.border },
              ]}
            >
              {selected ? (
                <View
                  style={[styles.dot, { backgroundColor: colors.primary }]}
                />
              ) : null}
            </View>
            <Text style={[styles.label, { color: colors.text }]}>{name}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  group: { gap: spacing.sm },
  option: {
    minHeight: minTouchTarget,
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    paddingHorizontal: spacing.md,
    borderWidth: 2,
    borderRadius: 8,
  },
  radio: {
    width: 22,
    height: 22,
    borderRadius: 11,
    borderWidth: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  dot: { width: 10, height: 10, borderRadius: 5 },
  label: { fontSize: fontSize.body },
});
