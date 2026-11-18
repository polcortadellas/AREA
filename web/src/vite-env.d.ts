/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Base URL of the AREA API, e.g. http://localhost:8080. */
  readonly VITE_API_URL?: string;
  /** "true" to answer API calls with MSW mocks during `npm run dev`. */
  readonly VITE_API_MOCKING?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
