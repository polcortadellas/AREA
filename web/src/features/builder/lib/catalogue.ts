import type {
  About,
  AboutNode,
  AboutService,
  Field,
} from '@/shared/api/generated/model';

export type UpstreamOutputOption = {
  token: string;
  reference: string;
  sourceId: string;
  sourceName: string;
  field: Field;
};

/** Services from /about.json that expose at least one trigger action. */
export function triggerServices(about: About): AboutService[] {
  return about.server.services.filter(service => service.actions.length > 0);
}

/** Services from /about.json that expose at least one reaction. */
export function reactionServices(about: About): AboutService[] {
  return about.server.services.filter(service => service.reactions.length > 0);
}

export function findServiceByName(
  about: About,
  serviceName: string,
): AboutService | undefined {
  return about.server.services.find(service => service.name === serviceName);
}

export function findServiceByNodeType(
  about: About,
  kind: 'trigger' | 'reaction',
  nodeType: string,
): AboutService | undefined {
  return about.server.services.find(service => {
    const nodes = kind === 'trigger' ? service.actions : service.reactions;
    return nodes.some(node => node.type === nodeType);
  });
}

export function findNodeByType(
  about: About,
  kind: 'trigger' | 'reaction',
  nodeType: string,
): AboutNode | undefined {
  for (const service of about.server.services) {
    const nodes = kind === 'trigger' ? service.actions : service.reactions;
    const match = nodes.find(node => node.type === nodeType);
    if (match) return match;
  }
  return undefined;
}

export function defaultInputValues(
  inputs: Field[],
  existing: Record<string, string | boolean> = {},
): Record<string, string | boolean> {
  const next: Record<string, string | boolean> = {};
  for (const field of inputs) {
    const current = existing[field.key];
    if (field.type === 'boolean') {
      next[field.key] = typeof current === 'boolean' ? current : false;
    } else {
      next[field.key] = typeof current === 'string' ? current : '';
    }
  }
  return next;
}

/**
 * Collects all declared output keys (`{{ trigger.key }}` and `{{ node.key }}`)
 * from ancestor nodes that execute before the current step.
 */
export function buildUpstreamOutputs(
  about: About,
  triggerType: string,
  ancestorReactions: ReadonlyArray<{ id: string; type: string }>,
): UpstreamOutputOption[] {
  const options: UpstreamOutputOption[] = [];
  const triggerNode = findNodeByType(about, 'trigger', triggerType);
  if (triggerNode) {
    for (const field of triggerNode.outputs) {
      options.push({
        token: `{{ trigger.${field.key} }}`,
        reference: `trigger.${field.key}`,
        sourceId: 'trigger',
        sourceName: triggerNode.type,
        field,
      });
    }
  }

  for (const ancestor of ancestorReactions) {
    const reactionNode = findNodeByType(about, 'reaction', ancestor.type);
    if (!reactionNode) continue;
    for (const field of reactionNode.outputs) {
      options.push({
        token: `{{ ${ancestor.id}.${field.key} }}`,
        reference: `${ancestor.id}.${field.key}`,
        sourceId: ancestor.id,
        sourceName: reactionNode.type,
        field,
      });
    }
  }

  return options;
}

/** Appends a `{{ node.key }}` token to an input value with clean spacing. */
export function appendKeyToken(current: string, token: string): string {
  const trimmed = current.trim();
  if (!trimmed) return token;
  return current.endsWith(' ') ? `${current}${token}` : `${current} ${token}`;
}
