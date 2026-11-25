import type {
  RunStep,
  Workflow,
  WorkflowNode,
} from '@/shared/api/generated/model';

/** Formats an ISO timestamp for the active UI language, returning null for missing/invalid values. */
export function formatDateTime(
  iso: string | null | undefined,
  locale: string,
): string | null {
  if (!iso) return null;
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return null;
  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

/** Formats elapsed time between two ISO timestamps (e.g. "450 ms", "1.45 s", "2m 05s"). */
export function formatDuration(
  startedAt: string | null | undefined,
  finishedAt: string | null | undefined,
): string | null {
  if (!startedAt || !finishedAt) return null;
  const startMs = Date.parse(startedAt);
  const endMs = Date.parse(finishedAt);
  if (Number.isNaN(startMs) || Number.isNaN(endMs) || endMs < startMs) {
    return null;
  }
  const diffMs = endMs - startMs;
  if (diffMs < 1000) return `${diffMs} ms`;
  if (diffMs < 60_000) return `${(diffMs / 1000).toFixed(2)} s`;
  const totalSeconds = Math.floor(diffMs / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}m ${String(seconds).padStart(2, '0')}s`;
}

/** Shortens a UUID for scannable list badges while keeping the full ID in title/details. */
export function shortId(id: string): string {
  return id.length > 8 ? id.slice(0, 8) : id;
}

/** Returns true when a nullable JSON object has at least one own property. */
export function hasRecordEntries(
  value: Record<string, unknown> | null | undefined,
): value is Record<string, unknown> {
  return value !== null && value !== undefined && Object.keys(value).length > 0;
}

/** Formats a JSON payload or step output with stable indentation. */
export function formatJson(value: unknown): string {
  return JSON.stringify(value, null, 2);
}

/**
 * Orders run steps in execution order:
 * 1. If the workflow graph is known, traverses from trigger along edges (true branch before false).
 * 2. Otherwise sorts executed steps by `started_at` ascending, placing unstarted/skipped steps after.
 */
export function orderRunSteps(
  steps: RunStep[],
  workflow?: Pick<Workflow, 'nodes' | 'edges'> | null,
): RunStep[] {
  if (steps.length <= 1) return steps;

  if (workflow && workflow.nodes.length > 0) {
    const graphOrder = computeGraphOrder(workflow.nodes, workflow.edges);
    if (graphOrder.size > 0) {
      return [...steps].sort((a, b) => {
        const idxA = graphOrder.get(a.node_id) ?? Number.MAX_SAFE_INTEGER;
        const idxB = graphOrder.get(b.node_id) ?? Number.MAX_SAFE_INTEGER;
        if (idxA !== idxB) return idxA - idxB;
        return compareByStartTime(a, b);
      });
    }
  }

  return [...steps].sort(compareByStartTime);
}

function compareByStartTime(a: RunStep, b: RunStep): number {
  const aTime = a.started_at
    ? Date.parse(a.started_at)
    : Number.POSITIVE_INFINITY;
  const bTime = b.started_at
    ? Date.parse(b.started_at)
    : Number.POSITIVE_INFINITY;
  if (aTime !== bTime) return aTime - bTime;
  return a.node_id.localeCompare(b.node_id);
}

function computeGraphOrder(
  nodes: WorkflowNode[],
  edges: Workflow['edges'],
): Map<string, number> {
  const order = new Map<string, number>();
  const trigger = nodes.find(n => n.kind === 'trigger');
  if (!trigger) return order;

  const outgoing = new Map<string, Workflow['edges']>();
  for (const edge of edges) {
    const list = outgoing.get(edge.from) ?? [];
    list.push(edge);
    outgoing.set(edge.from, list);
  }
  for (const list of outgoing.values()) {
    list.sort((a, b) => {
      if (a.branch === b.branch) return a.to.localeCompare(b.to);
      if (a.branch === 'true') return -1;
      if (b.branch === 'true') return 1;
      return (a.branch ?? '').localeCompare(b.branch ?? '');
    });
  }

  const visit = (id: string) => {
    if (order.has(id)) return;
    order.set(id, order.size);
    for (const edge of outgoing.get(id) ?? []) {
      visit(edge.to);
    }
  };

  visit(trigger.id);
  for (const node of nodes) {
    if (!order.has(node.id)) {
      order.set(node.id, order.size);
    }
  }
  return order;
}
