# ADR 0004: Typed connector nodes (TriggerNode / ReactionNode)

**Status:** Accepted
**Date:** 2026-11-15

## Context

Adding a service, Action or REAction must not change the execution engine, and its parameters and outputs must appear automatically in `/about.json` (ENG-EXT). Save-time validation of `{{ node.key }}` references also needs each node to declare what it outputs.

## Decision

Go interfaces `TriggerNode` (`ValidateConfig`, `InputSchema`, `OutputSchema`, `Poll`, `HandleWebhook`) and `ReactionNode` (`ValidateConfig`, `InputSchema`, `OutputSchema`, `Execute`), registered at start-up in a registry (`server/internal/node`, wired in `server/internal/connectors`). Triggers also declare an `InputSchema` because they take parameters too (a GitHub repository, a temperature threshold) that the builder must render and `/about.json` must expose. `/about.json` and the workflow validator are generated from that registry.

## Alternatives considered

- **Static JSON catalog of services:** rejected because it would be a second hard-coded catalogue that can drift from the code, and it cannot express validation or execution.

## Consequences

- **Positive:** a new connector touches only its own file and one registration line; the schemas drive `/about.json`, save-time validation (HTTP 422) and the web and mobile forms.
- **Negative (accepted):** every node must write and maintain its input and output schemas; connectors are compiled in, so adding one needs a rebuild and redeploy; no third-party code is loaded dynamically.
- **Not validated yet:** no spike covers this design. It is built in task M-09.
