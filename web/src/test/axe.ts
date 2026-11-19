import axe from 'axe-core';
import { expect } from 'vitest';

/**
 * Fails the test if axe-core finds WCAG A/AA violations in `container`.
 * color-contrast is skipped: jsdom does not compute styles, so it cannot be checked here.
 */
export async function expectNoA11yViolations(container: Element) {
  const { violations } = await axe.run(container, {
    runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa'] },
    rules: { 'color-contrast': { enabled: false } },
  });
  const summary = violations.map(
    v =>
      `${v.id}: ${v.help} (${v.nodes.map(n => n.target.join(' ')).join(', ')})`,
  );
  expect(summary).toEqual([]);
}
