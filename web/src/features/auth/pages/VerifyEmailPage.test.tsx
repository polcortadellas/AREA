import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { mockUser, mockVerifyToken } from '@/mocks/handlers/auth';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute } from '@/test/render';

describe('VerifyEmailPage', () => {
  it('confirms the email from the link and lands on the dashboard', async () => {
    const { user, container } = renderRoute(
      `/verify-email?token=${mockVerifyToken}`,
    );
    const button = await screen.findByRole('button', {
      name: 'Confirm my email',
    });
    await expectNoA11yViolations(container);

    await user.click(button);

    expect(
      await screen.findByRole('heading', { name: 'Your workflows' }),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(`Signed in as ${mockUser.email}`),
    ).toBeInTheDocument();
  });

  it('rejects an invalid or used link', async () => {
    const { user } = renderRoute('/verify-email?token=used-token');

    await user.click(
      await screen.findByRole('button', { name: 'Confirm my email' }),
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This link is invalid, already used or expired.',
    );
    expect(screen.getByRole('link', { name: 'Sign up again' })).toHaveAttribute(
      'href',
      '/register',
    );
  });

  it('handles a link without a token', async () => {
    renderRoute('/verify-email');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This link is incomplete.',
    );
    expect(
      screen.queryByRole('button', { name: 'Confirm my email' }),
    ).not.toBeInTheDocument();
  });
});
