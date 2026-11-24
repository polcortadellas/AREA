import type { RouteObject } from 'react-router';

import { GuestOnly } from './components/RouteGuards';

export const authRoutes: RouteObject[] = [
  {
    element: <GuestOnly />,
    children: [
      {
        path: 'login',
        lazy: async () => {
          const { LoginPage } = await import('./pages/LoginPage');
          return { Component: LoginPage };
        },
      },
      {
        path: 'register',
        lazy: async () => {
          const { RegisterPage } = await import('./pages/RegisterPage');
          return { Component: RegisterPage };
        },
      },
    ],
  },
  {
    path: 'verify-email',
    lazy: async () => {
      const { VerifyEmailPage } = await import('./pages/VerifyEmailPage');
      return { Component: VerifyEmailPage };
    },
  },
];
