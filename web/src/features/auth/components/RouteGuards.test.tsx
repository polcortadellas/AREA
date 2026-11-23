import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { mockUser } from '@/mocks/handlers/auth';
import { server } from '@/mocks/node';
import { renderRoute, signIn } from '@/test/render';

describe('session handling', () => {
  it('shows log in and sign up links, and no app navigation, when signed out', async () => {
    renderRoute('/login');

    await screen.findByRole('heading', { name: 'Log in' });
    const header = within(screen.getByRole('banner'));
    expect(header.getByRole('link', { name: 'Log in' })).toBeInTheDocument();
    expect(header.getByRole('link', { name: 'Sign up' })).toBeInTheDocument();
    expect(
      screen.queryByRole('navigation', { name: 'Main' }),
    ).not.toBeInTheDocument();
  });

  it('logs out and goes back to the login page', async () => {
    signIn();
    const { user } = renderRoute('/');
    await screen.findByText(`Signed in as ${mockUser.email}`);

    await user.click(screen.getByRole('button', { name: 'Log out' }));

    expect(
      await screen.findByRole('heading', { name: 'Log in' }),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole('banner')).getByRole('link', { name: 'Log in' }),
    ).toBeInTheDocument();
  });

  it('sends the user to log in when the session cookie has expired', async () => {
    server.use(
      http.get('*/auth/me', () =>
        HttpResponse.json({ error: 'unauthorized' }, { status: 401 }),
      ),
    );
    signIn();
    renderRoute('/');

    expect(
      await screen.findByRole('heading', { name: 'Log in' }),
    ).toBeInTheDocument();
  });
});
