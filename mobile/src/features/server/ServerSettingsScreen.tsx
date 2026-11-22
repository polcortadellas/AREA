import { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { useTranslation } from 'react-i18next';

import { Button } from '../../components/Button';
import { LanguageSwitcher } from '../../components/LanguageSwitcher';
import { Screen } from '../../components/Screen';
import { StatusMessage } from '../../components/StatusMessage';
import { TextField } from '../../components/TextField';
import { Body, Heading } from '../../components/Typography';
import type { RootScreenProps } from '../../navigation/types';
import { spacing } from '../../theme/theme';
import { useServerConfig } from './ServerConfigContext';
import {
  EMULATOR_HOST_URL,
  parseServerUrl,
  type ServerUrlError,
} from './serverUrl';
import { useConnectionCheck, useConnectionMessage } from './useConnectionCheck';

export function ServerSettingsScreen({
  navigation,
  route,
}: RootScreenProps<'Settings'>) {
  const { t } = useTranslation();
  const { serverUrl, saveServerUrl } = useServerConfig();
  const [input, setInput] = useState(serverUrl ?? '');
  const [error, setError] = useState<ServerUrlError | null>(null);
  const [saveState, setSaveState] = useState<'idle' | 'saved' | 'failed'>(
    'idle',
  );
  const { status, check, reset } = useConnectionCheck();
  const connection = useConnectionMessage(status);
  const firstRun = route.params?.firstRun ?? false;

  const validate = () => {
    const parsed = parseServerUrl(input);
    setError(parsed.ok ? null : parsed.error);
    return parsed.ok ? parsed.url : null;
  };

  const onChange = (text: string) => {
    setInput(text);
    setError(null);
    setSaveState('idle');
    reset();
  };

  const onTest = () => {
    const url = validate();
    if (url) {
      setInput(url);
      check(url);
    }
  };

  const onSave = async () => {
    const url = validate();
    if (!url) {
      return;
    }
    setInput(url);
    const ok = await saveServerUrl(url);
    setSaveState(ok ? 'saved' : 'failed');
    if (ok && firstRun) {
      navigation.reset({ index: 0, routes: [{ name: 'Home' }] });
    }
  };

  return (
    <Screen>
      <View style={styles.section}>
        <Heading>{t('settings.serverSection')}</Heading>
        <Body muted>{t('settings.serverIntro')}</Body>
        <TextField
          testID="server-url-input"
          label={t('settings.serverUrlLabel')}
          hint={t('settings.serverUrlHint')}
          error={error ? t(`validation.${error}`) : null}
          value={input}
          onChangeText={onChange}
          placeholder={EMULATOR_HOST_URL}
          autoCapitalize="none"
          autoCorrect={false}
          keyboardType="url"
          inputMode="url"
          returnKeyType="done"
          onSubmitEditing={onTest}
        />
        <View style={styles.actions}>
          <Button
            testID="server-test"
            variant="secondary"
            label={t('settings.test')}
            onPress={onTest}
            busy={status.state === 'checking'}
          />
          <Button
            testID="server-save"
            label={t('settings.save')}
            onPress={() => onSave()}
          />
        </View>
        {connection ? (
          <StatusMessage
            testID="server-status"
            tone={connection.tone}
            message={connection.message}
          />
        ) : null}
        {saveState !== 'idle' ? (
          <StatusMessage
            tone={saveState === 'saved' ? 'success' : 'error'}
            message={t(
              saveState === 'saved' ? 'settings.saved' : 'settings.saveFailed',
            )}
          />
        ) : null}
      </View>

      <View style={styles.section}>
        <Heading>{t('settings.languageSection')}</Heading>
        <LanguageSwitcher />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  section: { gap: spacing.md },
  actions: { gap: spacing.sm },
});
