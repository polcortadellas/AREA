import { resources } from './resources';

function keysOf(obj: object, prefix = ''): string[] {
  return Object.entries(obj).flatMap(([key, value]) =>
    typeof value === 'object' && value !== null
      ? keysOf(value, `${prefix}${key}.`)
      : [`${prefix}${key}`],
  );
}

describe('translations', () => {
  // The type check only catches keys missing from Spanish; this also catches extras.
  it('has exactly the same keys in English and Spanish', () => {
    expect(keysOf(resources.es.common).sort()).toEqual(
      keysOf(resources.en.common).sort(),
    );
  });

  it('has no empty strings', () => {
    for (const lng of ['en', 'es'] as const) {
      const empty = keysOf(resources[lng].common).filter(
        key =>
          key.split('.').reduce<any>((o, k) => o[k], resources[lng].common) ===
          '',
      );
      expect(empty).toEqual([]);
    }
  });
});
