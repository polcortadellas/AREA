import type { RouteObject } from 'react-router';

import { authRoutes, RequireAuth } from '@/features/auth';
import { builderRoutes } from '@/features/builder';
import { dashboardRoutes } from '@/features/dashboard';
import { servicesRoutes } from '@/features/services';
import { RootLayout } from './layouts/RootLayout';
import { NotFoundPage } from './pages/NotFoundPage';
import { RouteErrorPage } from './pages/RouteErrorPage';

/**
 * Each feature exports its own RouteObject[]; this file only composes them.
 * Kept separate from the router instance so tests can mount it in memory.
 */
export const routes: RouteObject[] = [
  {
    element: <RootLayout />,
    errorElement: <RouteErrorPage />,
    children: [
      ...authRoutes,
      {
        element: <RequireAuth />,
        children: [...dashboardRoutes, ...builderRoutes, ...servicesRoutes],
      },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
];
