# ADR 0005: Clients (React web and React Native Android)

**Status:** Accepted
**Date:** 2026-11-15 (evidence added 2026-11-24)

## Why a record

It decides two code bases and how they talk to the server. Switching later means rewriting every screen.

## Context

The subject asks for a web client and a mobile client that lists workflows, enables and disables them and shows what they did. Business logic stays on the server. Accessibility (WCAG AA) and security weigh heavily. See [`docs/tech-selection/frontend-selection.md`](../tech-selection/frontend-selection.md).

## Decision

React with Vite and TypeScript for the web, and React Native for Android, sharing the TypeScript types generated from the server's OpenAPI description.

## Alternatives considered

- **Vue (web):** scored higher in the matrix (60 against 52) but has no comparable mobile path, so the mobile client would need another technology.
- **Flutter (web and mobile):** one code base, but scored 36: the web build draws on a canvas, which is weak for accessibility, and the team does not know Dart.
- **Native Kotlin (Android):** a second language and a second tool chain for a client that only lists and toggles workflows.

## Consequences

- **Positive:** one language for both clients; the same OpenAPI-generated client; the team already has React experience.
- **Negative (accepted):** screens are written twice (the logic is shared, the components are not); React scored lower than Vue in the matrix, which has no mobile criterion; building the APK in Docker Compose is a risk (see the risk register in `planning/project-wbs.md`).

## Evidence

- Web: React 19, Vite, TypeScript and TanStack Query. The API client and the MSW mocks are generated from `openapi.yaml` with Orval ([`web/orval.config.ts`](../../web/orval.config.ts), [`web/src/shared/api/generated`](../../web/src/shared/api/generated)), M-17 and M-18.
- Accessibility: axe-core runs in the page tests, for example [`LoginPage.test.tsx`](../../web/src/features/auth/pages/LoginPage.test.tsx).
- Both clients use the same API: the web keeps the session in an HttpOnly cookie, the mobile client sends the bearer token (M-18).
- Android: React Native 0.88 with a configurable server address ([`mobile/src/features/server`](../../mobile/src/features/server)). The APK is built in Docker ([`mobile/Dockerfile`](../../mobile/Dockerfile)) and in CI when `mobile/` changes ([`android.yml`](../../.github/workflows/android.yml)), M-22. This removes the main risk named above.
- **Not validated yet:** the mobile client does not use the generated API client yet, so the shared types only benefit the web so far. Accessibility on Android has not been checked.

## Under the main alternative (Vue for the web)

- The web client would be Vue with its own generated client, and the Android client would need another technology (Kotlin or Flutter).
- That means two languages, two API clients and two tool chains for a team of four.
- The web forms built from `/about.json` would use Vue's two-way binding, which scored better for dynamic forms.
