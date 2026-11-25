import type { RouteObject } from 'react-router';

// The OAuth callback of the API redirects to /services, so the path is part of the contract.
export const servicesRoutes: RouteObject[] = [
  {
    path: 'services',
    lazy: async () => {
      const { ServicesPage } = await import('./pages/ServicesPage');
      return { Component: ServicesPage };
    },
  },
];
