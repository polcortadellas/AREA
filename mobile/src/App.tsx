import { useEffect } from 'react';
import { StatusBar } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { ServerConfigProvider } from './features/server';
import { restoreLanguage } from './i18n';
import { RootNavigator } from './navigation/RootNavigator';
import { useTheme } from './theme/theme';

function App() {
  const { dark } = useTheme();

  useEffect(() => {
    restoreLanguage();
  }, []);

  return (
    <SafeAreaProvider>
      <StatusBar barStyle={dark ? 'light-content' : 'dark-content'} />
      <ServerConfigProvider>
        <RootNavigator />
      </ServerConfigProvider>
    </SafeAreaProvider>
  );
}

export default App;
