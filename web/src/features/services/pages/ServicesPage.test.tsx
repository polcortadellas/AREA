import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { setMockSession } from '@/mocks/handlers/auth';
import { mockScopes, setMockLinks } from '@/mocks/handlers/services';
import { server } from '@/mocks/node';
import type { About } from '@/shared/api/generated/model';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';
import { goTo } from '../lib/redirect';

vi.mock('../lib/redirect', () => ({ goTo: vi.fn() }));

const about: About = {
  client: { host: '127.0.0.1' },
  server: {
    schema_version: 1,
    current_time: 1795217823,
    services: [
      {
        name: 'github',
        description: 'GitHub repositories and issues.',
        oauth2: true,
        actions: [],
        reactions: [],
      },
      {
        name: 'google',
        description: 'Gmail and Drive.',
        oauth2: true,
        actions: [],
        reactions: [],
      },
      {
        name: 'discord',
        description: 'Discord webhooks.',
        oauth2: false,
        actions: [],
        reactions: [],
      },
    ],
  },
};

const card = (service: string) =>
  screen.getByRole('heading', { name: service }).closest('li') as HTMLElement;

describe('ServicesPage', () => {
  beforeEach(() => {
    signIn();
    setMockLinks([]);
    vi.mocked(goTo).mockClear();
    server.use(http.get('*/about.json', () => HttpResponse.json(about)));
  });

  it('lists the services that need a linked account', async () => {
    const { container } = renderRoute('/services');

    expect(
      await screen.findByRole('heading', { name: 'github' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'google' })).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'discord' }),
    ).not.toBeInTheDocument();
    expect(within(card('github')).getByText('Not linked')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Services' })).toHaveAttribute(
      'href',
      '/services',
    );
    await expectNoA11yViolations(container);
  });

  it('sends the user to the provider to link a service', async () => {
    const { user } = renderRoute('/services');

    await user.click(
      await screen.findByRole('button', { name: 'Link github' }),
    );

    await vi.waitFor(() =>
      expect(goTo).toHaveBeenCalledWith(
        `${window.location.origin}/services?linked=github`,
      ),
    );
  });

  it('shows the result of the provider redirect', async () => {
    setMockLinks([
      {
        service: 'github',
        scopes: mockScopes,
        linked_at: '2026-11-24T10:00:00Z',
      },
    ]);
    const { container } = renderRoute('/services?linked=github');

    // "Loading…" is also a status, so look for the message itself.
    expect(await screen.findByText('github is now linked.')).toHaveAttribute(
      'role',
      'status',
    );
    await screen.findByRole('heading', { name: 'github' });
    const github = card('github');
    expect(
      within(github).getByText('Linked on Nov 24, 2026'),
    ).toBeInTheDocument();
    expect(
      within(github).getByText(`Permissions: ${mockScopes}`),
    ).toBeInTheDocument();
    expect(
      within(github).getByRole('button', { name: 'Unlink github' }),
    ).toBeInTheDocument();
    await expectNoA11yViolations(container);
  });

  it.each([
    ['access_denied', 'Linking was cancelled'],
    ['invalid_state', 'The link expired or was started in another browser'],
    ['link_failed', 'The service could not be linked'],
    ['something_new', 'The service could not be linked'],
  ])('explains the callback error %s', async (code, message) => {
    renderRoute(`/services?error=${code}`);

    expect(await screen.findByRole('alert')).toHaveTextContent(message);
  });

  it('unlinks a service', async () => {
    setMockLinks([
      {
        service: 'github',
        scopes: mockScopes,
        linked_at: '2026-11-24T10:00:00Z',
      },
    ]);
    const { user } = renderRoute('/services');

    await user.click(
      await screen.findByRole('button', { name: 'Unlink github' }),
    );

    expect(
      await screen.findByRole('button', { name: 'Link github' }),
    ).toBeInTheDocument();
    expect(within(card('github')).getByText('Not linked')).toBeInTheDocument();
  });

  it('says when the server has no OAuth app for a service', async () => {
    const { user } = renderRoute('/services');

    await user.click(
      await screen.findByRole('button', { name: 'Link google' }),
    );

    expect(await within(card('google')).findByRole('alert')).toHaveTextContent(
      'This service is not set up on this server yet.',
    );
    expect(goTo).not.toHaveBeenCalled();
  });

  it('sends a signed-out visitor to log in first', async () => {
    setMockSession(false);
    renderRoute('/services');

    expect(
      await screen.findByRole('heading', { name: 'Log in' }),
    ).toBeInTheDocument();
  });
});
