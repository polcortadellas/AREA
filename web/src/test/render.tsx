import type { ReactElement } from 'react';
import { QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createMemoryRouter, RouterProvider } from 'react-router';

import { routes } from '@/app/routes';
import { createQueryClient } from '@/shared/api/queryClient';

/** Renders the real route tree at `path`, with a fresh query cache per test. */
export function renderRoute(path = '/') {
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  return renderWithProviders(<RouterProvider router={router} />);
}

export function renderWithProviders(ui: ReactElement) {
  const queryClient = createQueryClient();
  queryClient.setDefaultOptions({ queries: { retry: false } });

  return {
    user: userEvent.setup(),
    ...render(
      <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
    ),
  };
}
