# AREA web client

React 19 + TypeScript single-page app built with Vite.

| Concern            | Library                             |
| ------------------ | ----------------------------------- |
| Routing            | React Router                        |
| Server state       | TanStack Query                      |
| UI                 | Tailwind CSS v4 + shadcn/ui (Radix) |
| Forms              | react-hook-form + zod               |
| Translations       | i18next (EN/ES)                     |
| Accessibility      | axe-core (dev console + tests)      |
| API mocks          | MSW                                 |
| API client codegen | Orval (from `openapi.yaml`)         |
| Tests              | Vitest + Testing Library            |

## Getting started

```bash
npm install
npm run dev        # http://localhost:5173, API calls answered by MSW mocks
```

| Script                                  | What it does                                             |
| --------------------------------------- | -------------------------------------------------------- |
| `npm run dev`                           | Dev server with hot reload                               |
| `npm run build`                         | Type-check and build to `dist/`                          |
| `npm test`                              | Run every test once (CI)                                 |
| `npm run test:watch`                    | Re-run tests on save                                     |
| `npm run lint` / `typecheck` / `format` | ESLint, `tsc`, Prettier (`make lint-web` runs all three) |
| `npm run api:gen`                       | Regenerate the API client and mocks from `openapi.yaml`  |

### Environment

Set in `.env.development` (committed) and overridden in `.env.development.local` (ignored by git):

| Variable           | Meaning                                                                              |
| ------------------ | ------------------------------------------------------------------------------------ |
| `VITE_API_URL`     | Base URL of the API, e.g. `http://localhost:8080`                                    |
| `VITE_API_MOCKING` | `true` to answer API calls with MSW in `npm run dev`; `false` to hit the real server |

The Docker image sets `VITE_API_URL` at build time (see `Dockerfile`). Mocks never run in production builds.

## Folder structure

```
src/
  app/              app shell: providers, route composition, layout, 404/error pages
  features/<name>/  one folder per feature (dashboard, auth, services, builder, ...)
    index.ts        the feature's public API (its routes, maybe a hook); import from here only
    routes.tsx      the feature's RouteObject[]
    pages/          route components
    components/     components used only by this feature
    hooks/          feature hooks (wrapping generated queries, local logic)
  shared/           code with no feature knowledge, usable everywhere
    api/            apiFetch, ApiError, QueryClient factory, generated/ (Orval output)
    ui/             shadcn components (Button, Input, Form, ...)
    i18n/           i18next setup, LanguageSwitcher, locales/{en,es}/*.json
    a11y/           dev-only axe checker
    lib/            small helpers (cn)
  mocks/            MSW handlers, shared by the dev worker and tests
  test/             test setup and helpers
```

Rules that keep responsibilities apart:

- `shared/` never imports from `features/` or `app/`.
- A feature never imports another feature's internals, only its `index.ts`.
- `app/routes.tsx` only composes the features' routes; it holds no page logic.
- Every server call goes through `shared/api/http.ts` (`apiFetch`), so auth headers and error handling live in one place.

### Adding a feature

1. Create `src/features/<name>/` with `pages/`, `routes.tsx` and `index.ts` exporting `<name>Routes`.
2. Spread `<name>Routes` into `children` in `src/app/routes.tsx`.
3. Add the text keys to both `locales/en/common.json` and `locales/es/common.json`.
4. Add a test next to the page (`*.test.tsx`) that renders it with `renderRoute('/path')` and calls `expectNoA11yViolations`.

## API client and mocks

The server contract lives in `openapi.yaml` (M-05). Orval reads it and writes to `src/shared/api/generated/`:

- request/response TypeScript types (`model/`)
- TanStack Query hooks, e.g. `useListWorkflows()`
- MSW handlers with faker data (`*.msw.ts`)

```bash
npm run api:gen                                   # default spec: ../server/api/openapi.yaml
OPENAPI_SPEC=path/to/openapi.yaml npm run api:gen # another location
```

After regenerating, register the new mock handlers in `src/mocks/handlers.ts` (e.g. `...getWorkflowsMock()`) and commit the generated folder: the web Docker build only sees `web/`, so it cannot read the spec from `server/`. The generated code is excluded from ESLint and Prettier; never edit it by hand.

`src/mocks/handlers/about.ts` and `src/mocks/handlers/workflows.ts` provide a deterministic `/about.json` catalogue and save-time workflow DAG + `{{ node.key }}` HTTP 422 validation mirroring `server/internal/node/validate.go`. Hand-written handlers listed before generated ones take priority, which is also how a test or a dev session overrides a single endpoint.

Errors: `apiFetch` throws `ApiError` with `status` and the parsed `body` (e.g. the HTTP 422 validation details), and generated hooks type their `error` as `ApiError`.

## Workflow builder (`src/features/builder/`)

Routes: `/workflows/new` (`POST /workflows`) and `/workflows/:workflowId/edit` (`GET` + `PUT /workflows/:workflowId`).

- **Zero hardcoded services:** `lib/catalogue.ts` derives trigger-capable services (`actions.length > 0`), reaction-capable services (`reactions.length > 0`), `oauth2` badges, parameter inputs (`string`, `number`, `boolean`) and declared outputs directly from `GET /about.json`.
- **DAG structure (`lib/workflowGraph.ts`):** serializes 1 trigger (`id: "start"`), an optional condition (`id: "condition"`, `type: "compare"` with `equal`, `not_equal`, `greater_than`, `less_than`) splitting into two non-merging branches (`branch: "true"` and `branch: "false"`), and 1 or more chained reactions per path.
- **Key mapping (`{{ trigger.key }}`, `{{ node.key }}`):** each condition and reaction step computes the declared outputs of all ancestor nodes on its execution path and exposes an accessible `Insert {{ node.key }}…` picker alongside direct template input.
- **Inline HTTP 422 validation (`lib/errors.ts`, `components/StepErrorAlert.tsx`):** `WorkflowValidationError.problems` returned by the server are matched by `problem.node` to the offending step card (`role="alert"`) and mark referenced input fields with `aria-invalid="true"`.

## Translations

- English (`locales/en/common.json`) is the source of truth. The Spanish file must have the same keys, or `tsc` fails.
- `t('key')` is type-checked: a misspelled key is a compile error.
- The language is detected from the browser, saved in `localStorage` (`area.lang`), and `<html lang>` follows it.
- Form validation: put an i18n key as the zod message (`z.email('validation.email')`); `FormMessage` translates it.

## Accessibility

- In `npm run dev`, axe-core scans the page after each change and logs WCAG A/AA violations to the browser console as `[axe]` warnings.
- In tests, `expectNoA11yViolations(container)` fails on any violation (colour contrast is skipped because jsdom has no styles).
- The layout has a skip link, a labelled main navigation and a focusable `<main>`. shadcn/Radix components handle keyboard and ARIA.
