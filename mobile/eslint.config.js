const { defineConfig, globalIgnores } = require('eslint/config');
const reactNative = require('@react-native/eslint-config/flat');

module.exports = defineConfig([
  globalIgnores(['android/', 'coverage/']),
  ...reactNative,
  {
    // We write TypeScript, not Flow. These Flow rules also crash on ESLint 9
    // (eslint-plugin-ft-flow still calls the removed context.getAllComments).
    rules: {
      'ft-flow/define-flow-type': 'off',
      'ft-flow/use-flow-type': 'off',
    },
  },
]);
