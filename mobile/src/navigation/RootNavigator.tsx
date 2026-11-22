import {
  DarkTheme,
  DefaultTheme,
  NavigationContainer,
} from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';

import { HomeScreen } from '../features/home/HomeScreen';
import { ServerSettingsScreen, useServerConfig } from '../features/server';
import { useTheme } from '../theme/theme';
import type { RootStackParamList } from './types';

const Stack = createNativeStackNavigator<RootStackParamList>();

export function RootNavigator() {
  const { t } = useTranslation();
  const { serverUrl, loaded } = useServerConfig();
  const { colors, dark } = useTheme();

  // Wait for the saved address so the first screen is the right one.
  if (!loaded) {
    return null;
  }

  const base = dark ? DarkTheme : DefaultTheme;
  const navTheme = {
    ...base,
    colors: {
      ...base.colors,
      primary: colors.primary,
      background: colors.background,
      card: colors.background,
      text: colors.text,
      border: colors.border,
    },
  };

  return (
    <NavigationContainer theme={navTheme}>
      <Stack.Navigator initialRouteName={serverUrl ? 'Home' : 'Settings'}>
        <Stack.Screen
          name="Home"
          component={HomeScreen}
          options={{ title: t('home.title') }}
        />
        <Stack.Screen
          name="Settings"
          component={ServerSettingsScreen}
          initialParams={serverUrl ? undefined : { firstRun: true }}
          options={({ route }) => ({
            title: t('settings.title'),
            headerBackVisible: !route.params?.firstRun,
          })}
        />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
