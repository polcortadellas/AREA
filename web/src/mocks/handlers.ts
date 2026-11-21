import { aboutHandlers } from './handlers/about';

/**
 * Every mock handler, shared by the browser worker (dev) and the Node server (tests).
 * After `npm run api:gen`, spread the generated ones in here, e.g.
 * `...getWorkflowsMock()` from '@/shared/api/generated/workflows/workflows.msw'.
 * Hand-written handlers listed first win over generated ones.
 */
export const handlers = [...aboutHandlers];
