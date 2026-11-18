import { defineConfig } from 'orval';

/**
 * Generates the typed API client, TanStack Query hooks and MSW mock handlers
 * from the server's OpenAPI contract (M-05). Run with `npm run api:gen`.
 * Override the spec location with OPENAPI_SPEC=path/to/openapi.yaml.
 */
const spec = process.env.OPENAPI_SPEC ?? '../server/api/openapi.yaml';

export default defineConfig({
  area: {
    input: { target: spec },
    output: {
      mode: 'tags-split',
      target: 'src/shared/api/generated',
      schemas: 'src/shared/api/generated/model',
      client: 'react-query',
      httpClient: 'fetch',
      mock: true,
      clean: true,
      prettier: true,
      override: {
        // Hooks return the response body directly instead of { data, status, headers }.
        fetch: { includeHttpResponseReturnType: false },
        mutator: {
          path: 'src/shared/api/http.ts',
          name: 'apiFetch',
        },
      },
    },
  },
});
