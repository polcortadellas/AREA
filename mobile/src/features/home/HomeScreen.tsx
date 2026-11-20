import { useEffect } from 'react';
import { StyleSheet, View } from 'react-native';
import { useTranslation } from 'react-i18next';

import { Button } from '../../components/Button';
import { Screen } from '../../components/Screen';
import { StatusMessage } from '../../components/StatusMessage';
import { Body, Heading } from '../../components/Typography';
import type { RootScreenProps } from '../../navigation/types';
import { spacing } from '../../theme/theme';
import {
  useConnectionCheck,
  useConnectionMessage,
  useServerConfig,
} from '../server';

export function HomeScreen({ navigation }: RootScreenProps<'Home'>) {
  const { t } = useTranslation();
  const { serverUrl } = useServerConfig();
  const { status, check } = useConnectionCheck();
  const connection = useConnectionMessage(status);

  useEffect(() => {
    if (serverUrl) {
      check(serverUrl);
    }
  }, [serverUrl, check]);

  return (
    <Screen>
      <View style={styles.section}>
        <Heading>{t('home.serverLabel')}</Heading>
        <Body testID="home-server-url" selectable>
          {serverUrl}
        </Body>
        {connection ? (
          <StatusMessage
            testID="home-status"
            tone={connection.tone}
            message={connection.message}
          />
        ) : null}
        {status.state === 'done' && !status.result.ok && serverUrl ? (
          <Button
            variant="secondary"
            label={t('home.retry')}
            onPress={() => check(serverUrl)}
          />
        ) : null}
        <Button
          testID="home-open-settings"
          label={t('home.openSettings')}
          onPress={() => navigation.navigate('Settings')}
        />
      </View>
      <Body muted>{t('home.workflowsPlaceholder')}</Body>
    </Screen>
  );
}

const styles = StyleSheet.create({
  section: { gap: spacing.md },
});
