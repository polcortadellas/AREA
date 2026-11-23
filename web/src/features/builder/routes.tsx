import type { RouteObject } from 'react-router';

import { WorkflowBuilderPage } from './pages/WorkflowBuilderPage';

export const builderRoutes: RouteObject[] = [
  {
    path: '/workflows/new',
    Component: WorkflowBuilderPage,
  },
  {
    path: '/workflows/:workflowId/edit',
    Component: WorkflowBuilderPage,
  },
];
