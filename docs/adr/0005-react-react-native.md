# ADR 0005: Clients (React web and React Native Android)

**Status:** Accepted
**Date:** 2026-11-15

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
- **Not validated yet:** React Native has no spike; it was chosen for sharing TypeScript with the web. Accessibility scores are estimates.
