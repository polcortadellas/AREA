import { screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';

describe('RootLayout', () => {
  beforeEach(signIn);

  it('renders the dashboard inside the layout', async () => {
    renderRoute('/');

    expect(
      await screen.findByRole('heading', { name: 'Your workflows' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('navigation', { name: 'Main' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('link', { name: 'Skip to main content' }),
    ).toHaveAttribute('href', '#main');
  });

  it('switches to Spanish without a reload and updates <html lang>', async () => {
    const { user } = renderRoute('/');
    await screen.findByRole('heading', { name: 'Your workflows' });

    await user.selectOptions(screen.getByLabelText('Language'), 'es');

    expect(
      await screen.findByRole('heading', { name: 'Tus flujos de trabajo' }),
    ).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('es');
    expect(window.localStorage.getItem('area.lang')).toBe('es');
  });

  it('has no accessibility violations', async () => {
    const { container } = renderRoute('/');
    await screen.findByRole('heading', { name: 'Your workflows' });

    await expectNoA11yViolations(container);
  });
});
