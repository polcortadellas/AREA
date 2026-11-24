import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import {
  mockPassword,
  mockUnverifiedEmail,
  mockUser,
} from '@/mocks/handlers/auth';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';

async function logIn(
  user: ReturnType<typeof renderRoute>['user'],
  email: string,
  password: string,
) {
  await user.type(await screen.findByLabelText('Email'), email);
  await user.type(screen.getByLabelText('Password'), password);
  await user.click(screen.getByRole('button', { name: 'Log in' }));
}

describe('LoginPage', () => {
  it('sends a signed-out visitor to log in, then to the dashboard', async () => {
    const { user, container } = renderRoute('/');

    expect(
      await screen.findByRole('heading', { name: 'Log in' }),
    ).toBeInTheDocument();
    await expectNoA11yViolations(container);

    await logIn(user, mockUser.email, mockPassword);

    expect(
      await screen.findByRole('heading', { name: 'Your workflows' }),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(`Signed in as ${mockUser.email}`),
    ).toBeInTheDocument();
  });

  it('validates the form before calling the server', async () => {
    const { user } = renderRoute('/login');

    await user.click(await screen.findByRole('button', { name: 'Log in' }));

    expect(
      await screen.findByText('Enter a valid email address'),
    ).toBeInTheDocument();
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });

  it('shows wrong credentials', async () => {
    const { user } = renderRoute('/login');

    await logIn(user, mockUser.email, 'wrong-password');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Wrong email or password.',
    );
  });

  it('explains that the email must be confirmed first', async () => {
    const { user } = renderRoute('/login');

    await logIn(user, mockUnverifiedEmail, mockPassword);

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Confirm your email first',
    );
  });

  it('returns to the requested page after logging in', async () => {
    const { user } = renderRoute('/login?next=%2Fdoes-not-exist');

    await logIn(user, mockUser.email, mockPassword);

    expect(
      await screen.findByRole('heading', { name: 'Page not found' }),
    ).toBeInTheDocument();
  });

  it('sends a logged-in user away from the login page', async () => {
    signIn();
    renderRoute('/login');

    expect(
      await screen.findByRole('heading', { name: 'Your workflows' }),
    ).toBeInTheDocument();
  });
});
