import { getAuthMock } from '@/shared/api/generated/auth/auth.msw';
import { getRunsMock } from '@/shared/api/generated/runs/runs.msw';
import { getServicesMock } from '@/shared/api/generated/services/services.msw';
import { getSystemMock } from '@/shared/api/generated/system/system.msw';
import { getWorkflowsMock } from '@/shared/api/generated/workflows/workflows.msw';
import { aboutHandlers } from './handlers/about';
import { authHandlers } from './handlers/auth';

/**
 * Every mock handler, shared by the browser worker (dev) and the Node server (tests).
 * Hand-written handlers listed first win over the generated ones (`npm run api:gen`).
 */
export const handlers = [
  ...aboutHandlers,
  ...authHandlers,
  ...getSystemMock(),
  ...getAuthMock(),
  ...getServicesMock(),
  ...getWorkflowsMock(),
  ...getRunsMock(),
];
