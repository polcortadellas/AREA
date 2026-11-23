import { describe, expect, it } from 'vitest';

import { safeNext } from './next';

describe('safeNext', () => {
  it.each([
    ['/workflows/1?tab=runs', '/workflows/1?tab=runs'],
    [null, '/'],
    ['', '/'],
    ['https://evil.example', '/'],
    ['//evil.example', '/'],
    ['/\\evil.example', '/'],
    ['javascript:alert(1)', '/'],
  ])('%s -> %s', (next, want) => {
    expect(safeNext(next)).toBe(want);
  });
});
