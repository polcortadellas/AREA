import { describe, expect, it } from 'vitest';

import type { RunStep, Workflow } from '@/shared/api/generated/model';
import {
  formatDateTime,
  formatDuration,
  hasRecordEntries,
  orderRunSteps,
  shortId,
} from './format';

describe('dashboard format helpers', () => {
  it('formats valid ISO timestamps and returns null for missing/invalid values', () => {
    expect(formatDateTime('2026-11-23T14:20:00.000Z', 'en')).toBeTruthy();
    expect(formatDateTime(null, 'en')).toBeNull();
    expect(formatDateTime('not-a-date', 'en')).toBeNull();
  });

  it('formats durations across milliseconds, seconds, and minutes', () => {
    expect(
      formatDuration('2026-11-23T14:00:00.000Z', '2026-11-23T14:00:00.450Z'),
    ).toBe('450 ms');
    expect(
      formatDuration('2026-11-23T14:00:00.000Z', '2026-11-23T14:00:01.450Z'),
    ).toBe('1.45 s');
    expect(
      formatDuration('2026-11-23T14:00:00.000Z', '2026-11-23T14:02:05.000Z'),
    ).toBe('2m 05s');
    expect(formatDuration(null, '2026-11-23T14:00:01.000Z')).toBeNull();
    expect(
      formatDuration('2026-11-23T14:00:02.000Z', '2026-11-23T14:00:01.000Z'),
    ).toBeNull();
  });

  it('shortens long IDs and detects non-empty records', () => {
    expect(shortId('11111111-1111-4111-8111-111111111111')).toBe('11111111');
    expect(shortId('abc')).toBe('abc');
    expect(hasRecordEntries({ count: 1 })).toBe(true);
    expect(hasRecordEntries({})).toBe(false);
    expect(hasRecordEntries(null)).toBe(false);
  });

  it('orders steps by workflow graph topology (true branch before false)', () => {
    const workflow: Pick<Workflow, 'nodes' | 'edges'> = {
      nodes: [
        { id: 't', kind: 'trigger', type: 'dummy.counter', config: {} },
        { id: 'c', kind: 'condition', type: 'condition', config: {} },
        { id: 'r_false', kind: 'reaction', type: 'dummy.echo', config: {} },
        { id: 'r_true', kind: 'reaction', type: 'dummy.echo', config: {} },
      ],
      edges: [
        { from: 't', to: 'c' },
        { from: 'c', to: 'r_false', branch: 'false' },
        { from: 'c', to: 'r_true', branch: 'true' },
      ],
    };

    const steps: RunStep[] = [
      {
        node_id: 'r_false',
        status: 'SKIPPED',
        attempt: 0,
        output: null,
        error: null,
        started_at: null,
        finished_at: null,
      },
      {
        node_id: 'r_true',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { ok: true },
        error: null,
        started_at: '2026-11-23T14:00:02.000Z',
        finished_at: '2026-11-23T14:00:03.000Z',
      },
      {
        node_id: 't',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { count: 1 },
        error: null,
        started_at: '2026-11-23T14:00:00.000Z',
        finished_at: '2026-11-23T14:00:01.000Z',
      },
      {
        node_id: 'c',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { result: true },
        error: null,
        started_at: '2026-11-23T14:00:01.000Z',
        finished_at: '2026-11-23T14:00:01.500Z',
      },
    ];

    expect(orderRunSteps(steps, workflow).map(s => s.node_id)).toEqual([
      't',
      'c',
      'r_true',
      'r_false',
    ]);
  });

  it('falls back to started_at ascending with skipped steps last when graph is absent', () => {
    const steps: RunStep[] = [
      {
        node_id: 'skipped_node',
        status: 'SKIPPED',
        attempt: 0,
        output: null,
        error: null,
        started_at: null,
        finished_at: null,
      },
      {
        node_id: 'second',
        status: 'SUCCEEDED',
        attempt: 1,
        output: null,
        error: null,
        started_at: '2026-11-23T14:00:02.000Z',
        finished_at: '2026-11-23T14:00:03.000Z',
      },
      {
        node_id: 'first',
        status: 'SUCCEEDED',
        attempt: 1,
        output: null,
        error: null,
        started_at: '2026-11-23T14:00:01.000Z',
        finished_at: '2026-11-23T14:00:01.500Z',
      },
    ];

    expect(orderRunSteps(steps).map(s => s.node_id)).toEqual([
      'first',
      'second',
      'skipped_node',
    ]);
  });
});
