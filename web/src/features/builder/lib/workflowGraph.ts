import type {
  About,
  Edge,
  Field,
  Workflow,
  WorkflowInput,
  WorkflowNode,
  WorkflowNodeConfig,
} from '@/shared/api/generated/model';
import {
  defaultInputValues,
  findNodeByType,
  findServiceByNodeType,
  reactionServices,
  triggerServices,
} from './catalogue';

export const CONDITION_OPERATORS = [
  'equal',
  'not_equal',
  'greater_than',
  'less_than',
] as const;

export type ConditionOperator = (typeof CONDITION_OPERATORS)[number];

export type DraftConfig = Record<string, string | boolean>;

export type DraftTrigger = {
  id: string;
  service: string;
  type: string;
  config: DraftConfig;
};

export type DraftCondition = {
  id: string;
  left: string;
  op: ConditionOperator;
  right: string;
};

export type DraftReaction = {
  id: string;
  service: string;
  type: string;
  config: DraftConfig;
};

export type WorkflowDraft = {
  name: string;
  enabled: boolean;
  trigger: DraftTrigger;
  condition: DraftCondition | null;
  reactions: DraftReaction[];
  trueBranchReactions: DraftReaction[];
  falseBranchReactions: DraftReaction[];
};

export function createDefaultReaction(about: About, id: string): DraftReaction {
  const firstService = reactionServices(about)[0];
  const firstReaction = firstService?.reactions[0];
  return {
    id,
    service: firstService?.name ?? '',
    type: firstReaction?.type ?? '',
    config: defaultInputValues(firstReaction?.inputs ?? []),
  };
}

export function createInitialDraft(about: About): WorkflowDraft {
  const firstTriggerService = triggerServices(about)[0];
  const firstAction = firstTriggerService?.actions[0];

  return {
    name: '',
    enabled: true,
    trigger: {
      id: 'start',
      service: firstTriggerService?.name ?? '',
      type: firstAction?.type ?? '',
      config: defaultInputValues(firstAction?.inputs ?? []),
    },
    condition: null,
    reactions: [createDefaultReaction(about, 'reaction_1')],
    trueBranchReactions: [createDefaultReaction(about, 'true_1')],
    falseBranchReactions: [createDefaultReaction(about, 'false_1')],
  };
}

export function nextReactionId(
  prefix: 'reaction' | 'true' | 'false',
  existing: ReadonlyArray<DraftReaction>,
): string {
  const used = new Set(existing.map(r => r.id));
  let index = existing.length + 1;
  while (used.has(`${prefix}_${index}`)) {
    index++;
  }
  return `${prefix}_${index}`;
}

export function activeNodeIds(draft: WorkflowDraft): Set<string> {
  const ids = new Set<string>([draft.trigger.id]);
  if (draft.condition) {
    ids.add(draft.condition.id);
    for (const r of draft.trueBranchReactions) ids.add(r.id);
    for (const r of draft.falseBranchReactions) ids.add(r.id);
  } else {
    for (const r of draft.reactions) ids.add(r.id);
  }
  return ids;
}

export function serializeConfig(
  inputs: Field[] | undefined,
  draftConfig: DraftConfig,
): WorkflowNodeConfig {
  const config: WorkflowNodeConfig = {};
  for (const field of inputs ?? []) {
    const raw = draftConfig[field.key];
    if (field.type === 'boolean') {
      config[field.key] = Boolean(raw);
      continue;
    }
    const text = typeof raw === 'string' ? raw : '';
    if (text.trim() === '') {
      continue;
    }
    if (field.type === 'number' && !text.includes('{{')) {
      const parsed = Number(text.trim());
      config[field.key] = Number.isNaN(parsed) ? text : parsed;
    } else {
      config[field.key] = text;
    }
  }
  return config;
}

export function toWorkflowInput(
  draft: WorkflowDraft,
  about: About,
): WorkflowInput {
  const nodes: WorkflowNode[] = [];
  const edges: Edge[] = [];

  const triggerDef = findNodeByType(about, 'trigger', draft.trigger.type);
  nodes.push({
    id: draft.trigger.id,
    kind: 'trigger',
    type: draft.trigger.type,
    config: serializeConfig(triggerDef?.inputs, draft.trigger.config),
  });

  if (!draft.condition) {
    let prevId = draft.trigger.id;
    for (const reaction of draft.reactions) {
      const def = findNodeByType(about, 'reaction', reaction.type);
      nodes.push({
        id: reaction.id,
        kind: 'reaction',
        type: reaction.type,
        config: serializeConfig(def?.inputs, reaction.config),
      });
      edges.push({ from: prevId, to: reaction.id });
      prevId = reaction.id;
    }
  } else {
    const condConfig: WorkflowNodeConfig = { op: draft.condition.op };
    if (draft.condition.left.trim() !== '') {
      condConfig.left = draft.condition.left;
    }
    if (draft.condition.right.trim() !== '') {
      condConfig.right = draft.condition.right;
    }
    nodes.push({
      id: draft.condition.id,
      kind: 'condition',
      type: 'compare',
      config: condConfig,
    });
    edges.push({ from: draft.trigger.id, to: draft.condition.id });

    const appendBranch = (
      branch: 'true' | 'false',
      branchReactions: ReadonlyArray<DraftReaction>,
    ) => {
      branchReactions.forEach((reaction, idx) => {
        const def = findNodeByType(about, 'reaction', reaction.type);
        nodes.push({
          id: reaction.id,
          kind: 'reaction',
          type: reaction.type,
          config: serializeConfig(def?.inputs, reaction.config),
        });
        if (idx === 0) {
          edges.push({
            from: draft.condition!.id,
            to: reaction.id,
            branch,
          });
        } else {
          edges.push({
            from: branchReactions[idx - 1].id,
            to: reaction.id,
          });
        }
      });
    };

    appendBranch('true', draft.trueBranchReactions);
    appendBranch('false', draft.falseBranchReactions);
  }

  return {
    name: draft.name.trim(),
    enabled: draft.enabled,
    nodes,
    edges,
  };
}

function hydrateConfig(
  inputs: Field[] | undefined,
  rawConfig: WorkflowNodeConfig,
): DraftConfig {
  const existing: Record<string, string | boolean> = {};
  for (const [k, v] of Object.entries(rawConfig ?? {})) {
    existing[k] = typeof v === 'boolean' ? v : String(v);
  }
  if (!inputs) return existing;
  return defaultInputValues(inputs, existing);
}

function toDraftReaction(node: WorkflowNode, about: About): DraftReaction {
  const service = findServiceByNodeType(about, 'reaction', node.type);
  const def = findNodeByType(about, 'reaction', node.type);
  return {
    id: node.id,
    service: service?.name ?? node.type.split('.')[0] ?? '',
    type: node.type,
    config: hydrateConfig(def?.inputs, node.config),
  };
}

export function fromWorkflow(workflow: Workflow, about: About): WorkflowDraft {
  const base = createInitialDraft(about);
  const nodeById = new Map(workflow.nodes.map(n => [n.id, n]));
  const triggerNode = workflow.nodes.find(n => n.kind === 'trigger');
  const conditionNode = workflow.nodes.find(n => n.kind === 'condition');

  const trigger: DraftTrigger = triggerNode
    ? {
        id: triggerNode.id,
        service:
          findServiceByNodeType(about, 'trigger', triggerNode.type)?.name ??
          triggerNode.type.split('.')[0] ??
          base.trigger.service,
        type: triggerNode.type,
        config: hydrateConfig(
          findNodeByType(about, 'trigger', triggerNode.type)?.inputs,
          triggerNode.config,
        ),
      }
    : base.trigger;

  const followChain = (startId: string | undefined): DraftReaction[] => {
    const chain: DraftReaction[] = [];
    const visited = new Set<string>();
    let currentId = startId;
    while (currentId && !visited.has(currentId)) {
      visited.add(currentId);
      const node = nodeById.get(currentId);
      if (!node || node.kind !== 'reaction') break;
      chain.push(toDraftReaction(node, about));
      const nextEdge = workflow.edges.find(e => e.from === currentId);
      currentId = nextEdge?.to;
    }
    return chain;
  };

  if (!conditionNode) {
    const firstReactionEdge = workflow.edges.find(e => e.from === trigger.id);
    const reactions = followChain(firstReactionEdge?.to);
    return {
      ...base,
      name: workflow.name,
      enabled: workflow.enabled,
      trigger,
      condition: null,
      reactions: reactions.length > 0 ? reactions : base.reactions,
    };
  }

  const rawOp = String(conditionNode.config?.op ?? 'equal');
  const op: ConditionOperator = (
    CONDITION_OPERATORS as readonly string[]
  ).includes(rawOp)
    ? (rawOp as ConditionOperator)
    : 'equal';

  const trueStart = workflow.edges.find(
    e => e.from === conditionNode.id && e.branch === 'true',
  )?.to;
  const falseStart = workflow.edges.find(
    e => e.from === conditionNode.id && e.branch === 'false',
  )?.to;

  return {
    ...base,
    name: workflow.name,
    enabled: workflow.enabled,
    trigger,
    condition: {
      id: conditionNode.id,
      left: String(conditionNode.config?.left ?? ''),
      op,
      right: String(conditionNode.config?.right ?? ''),
    },
    trueBranchReactions: followChain(trueStart),
    falseBranchReactions: followChain(falseStart),
  };
}
