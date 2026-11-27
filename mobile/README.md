# AREA Android client

React Native 0.88 + TypeScript app for Android. The server address is chosen at runtime, in the app, so a single APK works against any AREA server.

| Concern      | Library                                                |
| ------------ | ------------------------------------------------------ |
| Navigation   | React Navigation (native stack)                        |
| Translations | i18next + react-i18next (EN/ES), react-native-localize |
| Storage      | AsyncStorage (server address, chosen language)         |
| Tests        | Jest + React Native Testing Library                    |

## Getting started

Requirements: Node.js 22.11+, JDK 17, Android SDK (Android Studio) with an emulator or a device with USB debugging.

```bash
npm install
npm start          # Metro bundler
npm run android    # build a debug APK and install it on the running emulator/device
```

| Script                                  | What it does                                                |
| --------------------------------------- | ----------------------------------------------------------- |
| `npm start`                             | Metro bundler with fast refresh                             |
| `npm run android`                       | Debug build, installed on the connected emulator/device     |
| `npm test`                              | Run every test once                                         |
| `npm run lint` / `typecheck` / `format` | ESLint, `tsc`, Prettier (`make lint-mobile` runs all three) |

## Installing the APK from Docker Compose

`docker compose up --build` writes a release APK (arm64 for phones, x86_64 for emulators) to `./dist/AREA.apk`.

```bash
adb install -r dist/AREA.apk   # emulator or USB device
```

Or copy the file to the phone and open it (allow "Install unknown apps" for your file manager).

## Server address

On first launch the app asks for the server address; it is saved on the device and can be changed later from **Server settings**. **Test connection** calls `GET /ping` on that address and expects `pong`.

| Where the app runs                        | Address to enter                                                                                 |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------ |
| Android emulator, server on your computer | `http://10.0.2.2:8080` (the emulator's alias for the host)                                       |
| Phone on the same Wi-Fi                   | `http://<your computer's LAN IP>:8080`, e.g. `http://192.168.1.20:8080` (`ip addr` / `ipconfig`) |
| Deployed server                           | `https://area.example.com`                                                                       |

`http://` is assumed when no scheme is typed, so `192.168.1.20:8080` works too. The port is `API_PORT` from `.env` (8080 by default). If the phone cannot connect, check that the firewall on your computer allows incoming connections on that port.

Plain `http://` is allowed in release builds (`android/app/src/main/res/xml/network_security_config.xml`), because a self-hosted server on a LAN rarely has a certificate. Use `https://` for anything reachable from the internet.

## Accessibility and translations

- Shared components in `src/components/` set the accessibility role, label, state and hint, and have touch targets of at least 48dp. Colours (light and dark) meet WCAG AA; the contrast ratios are listed in `src/theme/theme.ts`.
- Errors and connection results are announced by TalkBack (live regions).
- The language follows the device (English or Spanish, English otherwise) until the user picks one in Settings; the choice is saved.
- Translations live in `src/i18n/locales/{en,es}/common.json`. English is the source of truth: `t('...')` keys are type-checked against it, and type-checking fails if Spanish is missing a key.

## Folder structure

```
index.js              registers the app; loads i18n first
src/
  App.tsx             providers (safe area, server config) and navigator
  navigation/         stack navigator and route param types
  features/<name>/    one folder per feature
    server/           server address: parsing, persistence, connection check, settings screen
    home/             home screen (workflow list comes with F-01)
  components/         accessible building blocks: Button, TextField, Heading/Body, StatusMessage, Screen, LanguageSwitcher
  i18n/               i18next setup and locales/{en,es}/common.json
  theme/              colour, spacing and size tokens
  storage/            AsyncStorage keys and safe read/write helpers
```
