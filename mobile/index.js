/**
 * @format
 */

// i18next must be initialised before the first component renders.
import './src/i18n';

import { AppRegistry } from 'react-native';
import App from './src/App';
import { name as appName } from './app.json';

AppRegistry.registerComponent(appName, () => App);
