import type { RouteObject } from 'react-router';

import { dashboardRoutes } from '@/features/dashboard';
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
    children: [...dashboardRoutes, { path: '*', element: <NotFoundPage /> }],
  },
];
