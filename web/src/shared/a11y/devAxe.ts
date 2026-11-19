import axe from 'axe-core';

const DEBOUNCE_MS = 1000;

/**
 * Dev-only accessibility checker. Runs axe-core on the page after the DOM
 * settles and logs each new WCAG A/AA violation to the browser console.
 * (@axe-core/react does not support React 18+, so this uses axe-core directly.)
 */
export function startDevAxe() {
  const reported = new Set<string>();
  let timer: number | undefined;
  let running = false;

  const run = async () => {
    if (running) return;
    running = true;
    try {
      const { violations } = await axe.run(document, {
        runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa'] },
      });
      for (const violation of violations) {
        for (const node of violation.nodes) {
          const key = `${violation.id}|${node.target.join(',')}`;
          if (reported.has(key)) continue;
          reported.add(key);
          console.warn(
            `[axe] ${violation.impact ?? 'unknown'}: ${violation.help}\n` +
              `  element: ${node.target.join(' ')}\n` +
              `  ${violation.helpUrl}`,
          );
        }
      }
    } finally {
      running = false;
    }
  };

  const schedule = () => {
    window.clearTimeout(timer);
    timer = window.setTimeout(() => void run(), DEBOUNCE_MS);
  };

  new MutationObserver(schedule).observe(document.body, {
    subtree: true,
    childList: true,
    attributes: true,
    characterData: true,
  });
  schedule();
}
