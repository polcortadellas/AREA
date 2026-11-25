import type { RouteObject } from 'react-router';

import { DashboardPage } from './pages/DashboardPage';
import { WorkflowHistoryPage } from './pages/WorkflowHistoryPage';

export const dashboardRoutes: RouteObject[] = [
  {
    index: true,
    Component: DashboardPage,
  },
  {
    path: 'workflows/:workflowId/runs',
    Component: WorkflowHistoryPage,
  },
  {
    path: 'workflows/:workflowId/runs/:runId',
    Component: WorkflowHistoryPage,
  },
];
