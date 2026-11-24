import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';

describe('NotFoundPage', () => {
  it('shows a translated 404 with a way back for unknown routes', async () => {
    signIn();
    const { container, user } = renderRoute('/does-not-exist');

    expect(
      screen.getByRole('heading', { name: 'Page not found' }),
    ).toBeInTheDocument();
    await expectNoA11yViolations(container);

    await user.click(screen.getByRole('link', { name: 'Back to dashboard' }));
    expect(
      await screen.findByRole('heading', { name: 'Your workflows' }),
    ).toBeInTheDocument();
  });
});
