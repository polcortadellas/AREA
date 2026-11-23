import type { RouteObject } from 'react-router';

export const builderRoutes: RouteObject[] = [
  {
    path: '/workflows/new',
    lazy: async () => {
      const { WorkflowBuilderPage } =
        await import('./pages/WorkflowBuilderPage');
      return { Component: WorkflowBuilderPage };
    },
  },
  {
    path: '/workflows/:workflowId/edit',
    lazy: async () => {
      const { WorkflowBuilderPage } =
        await import('./pages/WorkflowBuilderPage');
      return { Component: WorkflowBuilderPage };
    },
  },
];
