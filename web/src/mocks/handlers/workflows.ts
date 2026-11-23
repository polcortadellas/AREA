import { http, HttpResponse } from 'msw';

import type {
  AboutNode,
  Field,
  Workflow,
  WorkflowInput,
  WorkflowNode,
  WorkflowProblem,
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

let storedWorkflows: Workflow[] = [];

export function resetMockWorkflows(initial: Workflow[] = []) {
  storedWorkflows = [...initial];
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
  http.get('*/workflows', () => HttpResponse.json(storedWorkflows)),

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
      id: `wf-${storedWorkflows.length + 1}`,
      name: body.name || 'Untitled workflow',
      enabled: body.enabled ?? true,
      created_at: now,
      updated_at: now,
      nodes: body.nodes,
      edges: body.edges,
    };
    storedWorkflows = [created, ...storedWorkflows];
    return new HttpResponse(null, { status: 200 });
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
];
