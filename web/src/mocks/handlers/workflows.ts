import { http, HttpResponse, type PathParams } from 'msw';

import type {
  AboutNode,
  Field,
  Run,
  RunDetail,
  Workflow,
  WorkflowInput,
  WorkflowNode,
  WorkflowPatch,
  WorkflowProblem,
  WorkflowSummary,
} from '@/shared/api/generated/model';
import { mockAbout } from './about';

const CONDITION_OPS = ['equal', 'not_equal', 'greater_than', 'less_than'];
const CONDITION_SCHEMA: Field[] = [
  {
    key: 'left',
    type: 'string',
    required: true,
    description: 'Value to compare, usually a {{ node.key }} reference.',
  },
  {
    key: 'op',
    type: 'string',
    required: true,
    description: 'Comparison operator.',
  },
  {
    key: 'right',
    type: 'string',
    required: true,
    description: 'Value to compare against.',
  },
];

const REF_RE = /\{\{\s*([^{}\s.]+)\.([a-z][a-z0-9_]*)\s*\}\}/g;

const initialWorkflows: Workflow[] = [
  {
    id: '11111111-1111-4111-8111-111111111111',
    name: 'Notify Discord on new GitHub issue',
    enabled: true,
    created_at: '2026-11-20T09:15:00.000Z',
    updated_at: '2026-11-23T14:20:00.000Z',
    nodes: [
      {
        id: 'issue_trigger',
        kind: 'trigger',
        type: 'github.new_issue',
        config: { repo: 'polcortadellas/AREA' },
      },
      {
        id: 'is_bug',
        kind: 'condition',
        type: 'condition',
        config: {
          left: '{{ trigger.author }}',
          op: 'equal',
          right: 'polcortadellas',
        },
      },
      {
        id: 'discord_bug_alert',
        kind: 'reaction',
        type: 'discord.send_message',
        config: {
          webhook_url: 'https://discord.com/api/webhooks/123/abc',
          content: 'Issue #{{ trigger.number }}: {{ trigger.title }}',
        },
      },
      {
        id: 'echo_other',
        kind: 'reaction',
        type: 'dummy.echo',
        config: {
          message: 'Other author issue #{{ trigger.number }}',
        },
      },
    ],
    edges: [
      { from: 'issue_trigger', to: 'is_bug' },
      { from: 'is_bug', to: 'discord_bug_alert', branch: 'true' },
      { from: 'is_bug', to: 'echo_other', branch: 'false' },
    ],
  },
  {
    id: '22222222-2222-4222-8222-222222222222',
    name: 'Echo counter updates',
    enabled: false,
    created_at: '2026-11-19T11:00:00.000Z',
    updated_at: '2026-11-22T18:05:00.000Z',
    nodes: [
      {
        id: 'counter_start',
        kind: 'trigger',
        type: 'dummy.counter',
        config: {},
      },
      {
        id: 'echo_step',
        kind: 'reaction',
        type: 'dummy.echo',
        config: { message: 'Count reached {{ trigger.count }}' },
      },
    ],
    edges: [{ from: 'counter_start', to: 'echo_step' }],
  },
  {
    id: '33333333-3333-4333-8333-333333333333',
    name: 'Create Discord forum thread for releases',
    enabled: true,
    created_at: '2026-11-23T16:45:00.000Z',
    updated_at: '2026-11-23T16:45:00.000Z',
    nodes: [
      {
        id: 'release_trigger',
        kind: 'trigger',
        type: 'github.new_issue',
        config: { repo: 'polcortadellas/AREA' },
      },
      {
        id: 'forum_post',
        kind: 'reaction',
        type: 'discord.create_forum_thread',
        config: {
          webhook_url: 'https://discord.com/api/webhooks/456/def',
          thread_name: '{{ trigger.title }}',
          content: 'Release discussion for {{ trigger.title }}',
        },
      },
    ],
    edges: [{ from: 'release_trigger', to: 'forum_post' }],
  },
];

const initialRuns: RunDetail[] = [
  {
    id: 'aaaa1111-1111-4111-8111-111111111111',
    workflow_id: '11111111-1111-4111-8111-111111111111',
    status: 'SUCCEEDED',
    trigger_payload: {
      number: 42,
      title: 'Fix OAuth callback redirect',
      author: 'polcortadellas',
    },
    error: null,
    created_at: '2026-11-23T14:19:50.000Z',
    started_at: '2026-11-23T14:19:51.000Z',
    finished_at: '2026-11-23T14:19:52.450Z',
    steps: [
      {
        node_id: 'issue_trigger',
        status: 'SUCCEEDED',
        attempt: 1,
        output: {
          number: 42,
          title: 'Fix OAuth callback redirect',
          author: 'polcortadellas',
        },
        error: null,
        started_at: '2026-11-23T14:19:51.000Z',
        finished_at: '2026-11-23T14:19:51.050Z',
      },
      {
        node_id: 'is_bug',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { result: true },
        error: null,
        started_at: '2026-11-23T14:19:51.060Z',
        finished_at: '2026-11-23T14:19:51.075Z',
      },
      {
        node_id: 'discord_bug_alert',
        status: 'SUCCEEDED',
        attempt: 1,
        output: {
          message_id: '1309823412990',
          channel: '#bug-alerts',
          delivered: true,
        },
        error: null,
        started_at: '2026-11-23T14:19:51.100Z',
        finished_at: '2026-11-23T14:19:52.450Z',
      },
      {
        node_id: 'echo_other',
        status: 'SKIPPED',
        attempt: 0,
        output: null,
        error: null,
        started_at: null,
        finished_at: null,
      },
    ],
  },
  {
    id: 'aaaa2222-2222-4222-8222-222222222222',
    workflow_id: '11111111-1111-4111-8111-111111111111',
    status: 'FAILED',
    trigger_payload: {
      number: 41,
      title: 'Webhook timeout on deploy',
      author: 'polcortadellas',
    },
    error: 'node discord_bug_alert: discord webhook returned HTTP 429',
    created_at: '2026-11-23T12:04:10.000Z',
    started_at: '2026-11-23T12:04:11.000Z',
    finished_at: '2026-11-23T12:04:13.200Z',
    steps: [
      {
        node_id: 'issue_trigger',
        status: 'SUCCEEDED',
        attempt: 1,
        output: {
          number: 41,
          title: 'Webhook timeout on deploy',
          author: 'polcortadellas',
        },
        error: null,
        started_at: '2026-11-23T12:04:11.000Z',
        finished_at: '2026-11-23T12:04:11.030Z',
      },
      {
        node_id: 'is_bug',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { result: true },
        error: null,
        started_at: '2026-11-23T12:04:11.040Z',
        finished_at: '2026-11-23T12:04:11.055Z',
      },
      {
        node_id: 'discord_bug_alert',
        status: 'FAILED',
        attempt: 1,
        output: null,
        error: 'discord webhook returned HTTP 429: rate limited',
        started_at: '2026-11-23T12:04:11.070Z',
        finished_at: '2026-11-23T12:04:13.200Z',
      },
      {
        node_id: 'echo_other',
        status: 'SKIPPED',
        attempt: 0,
        output: null,
        error: null,
        started_at: null,
        finished_at: null,
      },
    ],
  },
  {
    id: 'bbbb1111-1111-4111-8111-111111111111',
    workflow_id: '22222222-2222-4222-8222-222222222222',
    status: 'SUCCEEDED',
    trigger_payload: { count: 7 },
    error: null,
    created_at: '2026-11-22T18:00:00.000Z',
    started_at: '2026-11-22T18:00:00.500Z',
    finished_at: '2026-11-22T18:00:00.780Z',
    steps: [
      {
        node_id: 'counter_start',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { count: 7 },
        error: null,
        started_at: '2026-11-22T18:00:00.500Z',
        finished_at: '2026-11-22T18:00:00.520Z',
      },
      {
        node_id: 'echo_step',
        status: 'SUCCEEDED',
        attempt: 1,
        output: { message: 'Count reached 7' },
        error: null,
        started_at: '2026-11-22T18:00:00.540Z',
        finished_at: '2026-11-22T18:00:00.780Z',
      },
    ],
  },
];

let storedWorkflows: Workflow[] = structuredClone(initialWorkflows);
let storedRuns: RunDetail[] = structuredClone(initialRuns);
let workflowSeq = 1;
let runSeq = 1;

export function resetMockWorkflows(
  initial: Workflow[] = initialWorkflows,
  runs: RunDetail[] = initialRuns,
) {
  storedWorkflows = structuredClone(initial);
  storedRuns = structuredClone(runs);
  workflowSeq = 1;
  runSeq = 1;
}

export function getMockWorkflows(): Workflow[] {
  return storedWorkflows;
}

function toSummary(wf: Workflow): WorkflowSummary {
  return {
    id: wf.id,
    name: wf.name,
    enabled: wf.enabled,
    created_at: wf.created_at,
    updated_at: wf.updated_at,
  };
}

function toRunSummary(run: RunDetail): Run {
  return {
    id: run.id,
    workflow_id: run.workflow_id,
    status: run.status,
    trigger_payload: run.trigger_payload,
    error: run.error,
    created_at: run.created_at,
    started_at: run.started_at,
    finished_at: run.finished_at,
  };
}

function findCatalogNode(
  kind: 'trigger' | 'reaction',
  type: string,
): AboutNode | undefined {
  for (const service of mockAbout.server.services) {
    const list = kind === 'trigger' ? service.actions : service.reactions;
    const found = list.find(n => n.type === type);
    if (found) return found;
  }
  return undefined;
}

function validateSchema(
  schema: Field[],
  values: Record<string, unknown>,
): string | null {
  const problems: string[] = [];
  for (const field of schema) {
    const v = values[field.key];
    if (v === undefined || v === null || v === '') {
      if (field.required) {
        problems.push(`"${field.key}" is required`);
      }
    } else if (typeof v === 'string' && v.includes('{{')) {
      // Template references are checked by checkRefs.
    } else if (typeof v !== field.type) {
      problems.push(`"${field.key}" must be a ${field.type}`);
    }
  }
  for (const key of Object.keys(values).sort()) {
    if (!schema.some(f => f.key === key)) {
      problems.push(`"${key}" is not a parameter of this node`);
    }
  }
  return problems.length > 0
    ? `invalid node configuration: ${problems.join('; ')}`
    : null;
}

function isAncestor(
  ref: string,
  id: string,
  parent: Map<string, string>,
): boolean {
  const seen = new Set<string>();
  let cur = parent.get(id);
  while (cur !== undefined && !seen.has(cur)) {
    if (cur === ref) return true;
    seen.add(cur);
    cur = parent.get(cur);
  }
  return false;
}

function outputSchemaOf(node: WorkflowNode): Field[] {
  if (node.kind === 'trigger' || node.kind === 'reaction') {
    return findCatalogNode(node.kind, node.type)?.outputs ?? [];
  }
  return [];
}

/**
 * Mirrors `server/internal/node/validate.go` so `npm run dev` and tests return
 * the same HTTP 422 `invalid_workflow` problems as the Go backend.
 */
export function validateMockWorkflow(input: WorkflowInput): WorkflowProblem[] {
  const problems: WorkflowProblem[] = [];
  const bad = (node: string | undefined, message: string) => {
    problems.push(node ? { node, message } : { message });
  };

  const nodes = new Map<string, WorkflowNode>();
  let triggerId = '';
  let conditions = 0;

  for (const n of input.nodes ?? []) {
    if (!n.id || n.id === 'trigger') {
      bad(undefined, `node id "${n.id}" is not allowed`);
      continue;
    }
    if (nodes.has(n.id)) {
      bad(n.id, 'node appears twice');
      continue;
    }
    nodes.set(n.id, n);

    const config = (n.config ?? {}) as Record<string, unknown>;
    if (n.kind === 'trigger') {
      if (triggerId !== '') {
        bad(n.id, 'a workflow has exactly one trigger');
      }
      triggerId = n.id;
      const trig = findCatalogNode('trigger', n.type);
      if (!trig) {
        bad(n.id, `unknown trigger "${n.type}"`);
      } else {
        const err = validateSchema(trig.inputs, config);
        if (err) bad(n.id, err);
      }
    } else if (n.kind === 'reaction') {
      const react = findCatalogNode('reaction', n.type);
      if (!react) {
        bad(n.id, `unknown reaction "${n.type}"`);
      } else {
        const err = validateSchema(react.inputs, config);
        if (err) bad(n.id, err);
      }
    } else if (n.kind === 'condition') {
      conditions++;
      const err = validateSchema(CONDITION_SCHEMA, config);
      if (err) {
        bad(n.id, err);
      } else if (
        typeof config.op !== 'string' ||
        !CONDITION_OPS.includes(config.op)
      ) {
        bad(
          n.id,
          `unknown operator "${String(config.op ?? '')}", want one of ${CONDITION_OPS.join(', ')}`,
        );
      }
    } else {
      bad(n.id, `unknown kind "${String(n.kind)}"`);
    }
  }

  if (triggerId === '') {
    bad(undefined, 'a workflow needs a trigger');
  }
  if (conditions > 1) {
    bad(undefined, `a workflow has at most one condition, found ${conditions}`);
  }

  const parent = new Map<string, string>();
  const children = new Map<string, string[]>();
  const branches = new Map<string, Set<string>>();

  for (const e of input.edges ?? []) {
    const from = nodes.get(e.from);
    const to = nodes.get(e.to);
    if (!from || !to) {
      bad(undefined, `edge ${e.from} -> ${e.to} points to a missing node`);
      continue;
    }
    if (e.to === triggerId) {
      bad(
        e.from,
        `edge ${e.from} -> ${e.to}: nothing can point to the trigger`,
      );
      continue;
    }
    if (
      from.kind === 'condition' &&
      e.branch !== 'true' &&
      e.branch !== 'false'
    ) {
      bad(
        e.from,
        `edge ${e.from} -> ${e.to}: an edge leaving a condition needs branch "true" or "false"`,
      );
      continue;
    }
    if (from.kind !== 'condition' && e.branch) {
      bad(e.from, `edge ${e.from} -> ${e.to}: only a condition has branches`);
      continue;
    }
    if (parent.has(e.to)) {
      bad(e.to, 'several incoming edges: branches cannot reconverge');
      continue;
    }
    if (from.kind === 'condition' && e.branch) {
      const seenBranches = branches.get(e.from) ?? new Set<string>();
      if (seenBranches.has(e.branch)) {
        bad(e.from, `branch "${e.branch}" has more than one edge`);
        continue;
      }
      seenBranches.add(e.branch);
      branches.set(e.from, seenBranches);
    }
    parent.set(e.to, e.from);
    children.set(e.from, [...(children.get(e.from) ?? []), e.to]);
  }

  for (const n of input.nodes ?? []) {
    if (n.kind === 'condition') {
      const seenBranches = branches.get(n.id) ?? new Set<string>();
      for (const b of ['true', 'false']) {
        if (!seenBranches.has(b)) {
          bad(n.id, `branch "${b}" is empty`);
        }
      }
    }
  }

  const reached = new Set<string>();
  const stack = triggerId ? [triggerId] : [];
  while (stack.length > 0) {
    const id = stack.pop();
    if (!id || reached.has(id)) continue;
    reached.add(id);
    stack.push(...(children.get(id) ?? []));
  }

  for (const n of input.nodes ?? []) {
    if (!reached.has(n.id)) {
      bad(n.id, 'not connected to the trigger');
      continue;
    }
    const config = (n.config ?? {}) as Record<string, unknown>;
    for (const key of Object.keys(config).sort()) {
      const val = config[key];
      if (typeof val !== 'string') continue;
      if (val.replace(REF_RE, '').includes('{{')) {
        bad(n.id, `"${key}" has a malformed {{ node.key }} reference`);
      }
      for (const match of val.matchAll(REF_RE)) {
        const rawRef = match[1];
        const field = match[2];
        const refId = rawRef === 'trigger' ? triggerId : rawRef;
        if (!isAncestor(refId, n.id, parent)) {
          bad(
            n.id,
            `"${key}": {{ ${rawRef}.${field} }} must point to a node that runs before this one`,
          );
          continue;
        }
        const refNode = nodes.get(refId);
        if (
          !refNode ||
          !outputSchemaOf(refNode).some(out => out.key === field)
        ) {
          bad(
            n.id,
            `"${key}": node "${rawRef}" does not declare output "${field}"`,
          );
        }
      }
    }
  }

  return problems;
}

export const workflowHandlers = [
  http.get('*/workflows', () => {
    const sorted = [...storedWorkflows]
      .sort((a, b) => b.created_at.localeCompare(a.created_at))
      .map(toSummary);
    return HttpResponse.json(sorted);
  }),

  http.post('*/workflows', async ({ request }) => {
    let body: WorkflowInput;
    try {
      body = (await request.json()) as WorkflowInput;
    } catch {
      return HttpResponse.json({ error: 'invalid_body' }, { status: 400 });
    }
    if (!body || typeof body !== 'object' || !Array.isArray(body.nodes)) {
      return HttpResponse.json({ error: 'invalid_body' }, { status: 400 });
    }

    const problems = validateMockWorkflow(body);
    if (problems.length > 0) {
      return HttpResponse.json(
        { error: 'invalid_workflow', problems },
        { status: 422 },
      );
    }

    const now = new Date().toISOString();
    const created: Workflow = {
      id: `wf-${workflowSeq++}`,
      name: body.name || 'Untitled workflow',
      enabled: body.enabled ?? true,
      created_at: now,
      updated_at: now,
      nodes: body.nodes,
      edges: body.edges,
    };
    storedWorkflows = [created, ...storedWorkflows];
    return HttpResponse.json(created, { status: 201 });
  }),

  http.get('*/workflows/:workflowId', ({ params }) => {
    const found = storedWorkflows.find(w => w.id === params.workflowId);
    if (!found) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    return HttpResponse.json(found);
  }),

  http.put('*/workflows/:workflowId', async ({ params, request }) => {
    const idx = storedWorkflows.findIndex(w => w.id === params.workflowId);
    if (idx === -1) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    let body: WorkflowInput;
    try {
      body = (await request.json()) as WorkflowInput;
    } catch {
      return HttpResponse.json({ error: 'invalid_body' }, { status: 400 });
    }

    const problems = validateMockWorkflow(body);
    if (problems.length > 0) {
      return HttpResponse.json(
        { error: 'invalid_workflow', problems },
        { status: 422 },
      );
    }

    const updated: Workflow = {
      ...storedWorkflows[idx],
      name: body.name,
      enabled: body.enabled ?? storedWorkflows[idx].enabled,
      updated_at: new Date().toISOString(),
      nodes: body.nodes,
      edges: body.edges,
    };
    storedWorkflows[idx] = updated;
    return HttpResponse.json(updated);
  }),

  http.patch<PathParams, WorkflowPatch>(
    '*/workflows/:workflowId',
    async ({ params, request }) => {
      const idx = storedWorkflows.findIndex(w => w.id === params.workflowId);
      if (idx === -1) {
        return HttpResponse.json({ error: 'not_found' }, { status: 404 });
      }
      let patch: WorkflowPatch;
      try {
        patch = await request.json();
      } catch {
        return HttpResponse.json({ error: 'invalid_body' }, { status: 400 });
      }
      if (!patch || (patch.name === undefined && patch.enabled === undefined)) {
        return HttpResponse.json({ error: 'invalid_body' }, { status: 400 });
      }
      const updated: Workflow = {
        ...storedWorkflows[idx],
        ...(patch.name !== undefined ? { name: patch.name } : {}),
        ...(patch.enabled !== undefined ? { enabled: patch.enabled } : {}),
        updated_at: new Date().toISOString(),
      };
      storedWorkflows[idx] = updated;
      return HttpResponse.json(updated);
    },
  ),

  http.delete('*/workflows/:workflowId', ({ params }) => {
    const exists = storedWorkflows.some(w => w.id === params.workflowId);
    if (!exists) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    storedWorkflows = storedWorkflows.filter(w => w.id !== params.workflowId);
    storedRuns = storedRuns.filter(r => r.workflow_id !== params.workflowId);
    return new HttpResponse(null, { status: 204 });
  }),

  http.post('*/workflows/:workflowId/trigger', ({ params }) => {
    const wf = storedWorkflows.find(w => w.id === params.workflowId);
    if (!wf) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    const now = new Date().toISOString();
    const id = `cccc0000-0000-4000-8000-${String(runSeq++).padStart(12, '0')}`;
    const triggerNode = wf.nodes.find(n => n.kind === 'trigger');
    const created: RunDetail = {
      id,
      workflow_id: wf.id,
      status: 'SUCCEEDED',
      trigger_payload: { manual: true },
      error: null,
      created_at: now,
      started_at: now,
      finished_at: now,
      steps: wf.nodes.map(node => ({
        node_id: node.id,
        status: 'SUCCEEDED',
        attempt: 1,
        output:
          node.id === triggerNode?.id
            ? { manual: true }
            : { ok: true, node: node.type },
        error: null,
        started_at: now,
        finished_at: now,
      })),
    };
    storedRuns = [created, ...storedRuns];
    return HttpResponse.json({ id }, { status: 202 });
  }),

  http.get('*/workflows/:workflowId/runs', ({ params, request }) => {
    const exists = storedWorkflows.some(w => w.id === params.workflowId);
    if (!exists) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    const url = new URL(request.url);
    const rawLimit = Number(url.searchParams.get('limit') ?? '20');
    const limit = Number.isFinite(rawLimit)
      ? Math.min(100, Math.max(1, rawLimit))
      : 20;
    const list = storedRuns
      .filter(r => r.workflow_id === params.workflowId)
      .sort((a, b) => b.created_at.localeCompare(a.created_at))
      .slice(0, limit)
      .map(toRunSummary);
    return HttpResponse.json(list);
  }),

  http.get('*/runs/:runId', ({ params }) => {
    const found = storedRuns.find(r => r.id === params.runId);
    if (!found) {
      return HttpResponse.json({ error: 'not_found' }, { status: 404 });
    }
    return HttpResponse.json(found);
  }),
];
