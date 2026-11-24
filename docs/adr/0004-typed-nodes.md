# ADR 0004: Typed connector nodes (TriggerNode / ReactionNode)

**Status:** Accepted
**Date:** 2026-11-15 (evidence added 2026-11-24)

## Why a record

It is the extension point of the product (ENG-EXT). Every connector the team writes follows it, so changing it later touches all of them.

## Context

Adding a service, Action or REAction must not change the execution engine, and its parameters and outputs must appear automatically in `/about.json` (ENG-EXT). Save-time validation of `{{ node.key }}` references also needs each node to declare what it outputs.

## Decision

Go interfaces `TriggerNode` (`ValidateConfig`, `InputSchema`, `OutputSchema`, `Poll`, `HandleWebhook`) and `ReactionNode` (`ValidateConfig`, `InputSchema`, `OutputSchema`, `Execute`), registered at start-up in a registry (`server/internal/node`, wired in `server/internal/connectors`). Triggers also declare an `InputSchema` because they take parameters too (a GitHub repository, a temperature threshold) that the builder must render and `/about.json` must expose. `/about.json` and the workflow validator are generated from that registry.

## Alternatives considered

- **Static JSON catalog of services:** rejected because it would be a second hard-coded catalogue that can drift from the code, and it cannot express validation or execution.

## Consequences

- **Positive:** a new connector touches only its own file and one registration line; the schemas drive `/about.json`, save-time validation (HTTP 422) and the web and mobile forms.
- **Negative (accepted):** every node must write and maintain its input and output schemas; connectors are compiled in, so adding one needs a rebuild and redeploy; no third-party code is loaded dynamically.

## Evidence

- Interfaces, schemas and registry in [`server/internal/node`](../../server/internal/node) (M-09). Registering an invalid or duplicated node panics at start-up, so the server never runs with a broken catalogue.
- Connectors are registered in one function, [`server/internal/connectors/connectors.go`](../../server/internal/connectors/connectors.go); the `dummy` connector shows the pattern.
- `/about.json` is built from the registry on every request ([`server/internal/api/about.go`](../../server/internal/api/about.go), M-06). `TestAboutListsNewNode` defines a node only inside the test and checks that it appears with its inputs and outputs.
- Save-time validation uses the same schemas: [`server/internal/node/validate.go`](../../server/internal/node/validate.go) rejects unknown types, bad config and undeclared `{{ node.key }}` references, answered with 422 (M-10).
- The executor looks nodes up in the same registry ([`server/internal/engine/engine.go`](../../server/internal/engine/engine.go), M-12).
- **Found in practice:** conditions are not registry nodes. Their schema and operators are fixed in `validate.go`, so a new operator needs a change there.

## Under the main alternative (static JSON catalogue)

- `/about.json` would be a file edited by hand next to the Go code, and a mismatch between the two would only show at run time.
- Checking a node's config and its `{{ }}` references would mean reading that file instead of calling `ValidateConfig`.
- Each service would still need Go code to run, so every connector would live in two places.
