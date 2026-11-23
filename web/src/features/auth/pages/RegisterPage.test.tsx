import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { server } from '@/mocks/node';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute } from '@/test/render';

async function fillForm(
  user: ReturnType<typeof renderRoute>['user'],
  password: string,
  confirm = password,
) {
  await user.type(await screen.findByLabelText('Email'), 'new@example.com');
  await user.type(screen.getByLabelText('Password'), password);
  await user.type(screen.getByLabelText('Confirm password'), confirm);
  await user.click(screen.getByRole('button', { name: 'Sign up' }));
}

describe('RegisterPage', () => {
  it('signs up and tells the user to check their email', async () => {
    const { user, container } = renderRoute('/register');
    await screen.findByRole('heading', { name: 'Create your account' });
    await expectNoA11yViolations(container);

    await fillForm(user, 'correct-horse');

    expect(
      await screen.findByRole('heading', { name: 'Check your email' }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/We sent a confirmation link to new@example.com/),
    ).toBeInTheDocument();
    await expectNoA11yViolations(container);
  });

  it('checks the password length and confirmation', async () => {
    const { user } = renderRoute('/register');

    await fillForm(user, 'short', 'different');

    expect(
      await screen.findByText('Use 8 to 128 characters'),
    ).toBeInTheDocument();
    expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
  });

  it('shows errors returned by the server', async () => {
    server.use(
      http.post('*/auth/register', () =>
        HttpResponse.json({ error: 'invalid_email' }, { status: 400 }),
      ),
    );
    const { user } = renderRoute('/register');

    await fillForm(user, 'correct-horse');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Enter a valid email address.',
    );
  });

  it('reports when the server cannot be reached', async () => {
    server.use(http.post('*/auth/register', () => HttpResponse.error()));
    const { user } = renderRoute('/register');

    await fillForm(user, 'correct-horse');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Cannot reach the server.',
    );
  });
});
