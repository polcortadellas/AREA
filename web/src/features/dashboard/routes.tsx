import type { RouteObject } from 'react-router';

export const dashboardRoutes: RouteObject[] = [
  {
    index: true,
    lazy: async () => {
      const { DashboardPage } = await import('./pages/DashboardPage');
      return { Component: DashboardPage };
    },
  },
];
