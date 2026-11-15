# Spike: Flutter Web ping page

## Goal

Check that a Flutter front-end can call the `/ping` endpoint of a backend (Hono/Bun in the original spike) and show the result, to compare it with the Vue and React spikes.

Recovered from `spikes/hono_bun/frontend` (commit `4a0cdf1`), which was removed from the tree when the Hono spike was cleaned up.

## What the code does

- `lib/main.dart`: one screen that calls `GET $API_URL/ping`, parses `{"status":"ok"}`, shows "Backend Connected" or "Backend Unreachable", the raw body or the error, and a button to ping again. 5 s timeout.
- The API address is not in the source: it comes from `--dart-define=API_URL=...` at build time.
- `test/widget_test.dart`: one widget test (the screen renders with the "Ping /ping" button).
- `Dockerfile`: multi-stage, builds with `ghcr.io/cirruslabs/flutter:stable`, serves the static web build with `nginx:alpine`.
- Dependencies: `http` only. The generated platform folders (`android/`, `ios/`, `linux/`, `macos/`, `windows/`) and the template README were dropped to keep the spike small; only `web/` is kept. `flutter create .` recreates them.

## What was NOT checked

No report, timings or measurements were kept with the original spike, and nothing was run when it was recovered. So:

- Not re-run: the test, the Docker build, the page in a browser.
- No image size, no bundle size, no build time.
- No experience notes (the team has no Dart experience on record).

## Strengths and weaknesses (from the code and documentation, not measured)

| Criterion | Note |
|---|---|
| Security | Flutter Web draws on a canvas, so there is no HTML templating to inject into. Token storage and CSRF are still up to us. |
| Dynamic forms | Widgets are built in Dart code, so a `/about.json` form is a `map()` into widgets. No two-way binding like Vue's `v-model`. |
| Accessibility & i18n | The weak point: the web build is canvas-based and needs the semantics tree to expose elements to screen readers. `flutter_localizations` covers i18n. |
| Learning curve | Dart and Flutter are new for the team. |
| OpenAPI / typing | Dart is strongly typed, but there is no TypeScript. Client generation would use `openapi-generator` (Dart). |
| Docker | The final image is small (nginx + static files), but the build image is large and the web bundle is heavier than Vue or React. |
| Extra | The same code base could produce the Android client the project asks for. This is the main reason to keep Flutter in the comparison. |

## How to reproduce

```bash
cd spikes/flutter
flutter pub get
flutter test
flutter run -d chrome --dart-define=API_URL=http://localhost:8080

# Docker
docker build --build-arg API_URL=http://localhost:8080 -t area-spike-flutter .
docker run --rm -p 8081:80 area-spike-flutter
```
