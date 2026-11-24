import type { WorkflowValidationError } from '@/shared/api/generated/model';
import { ApiError } from '@/shared/api/http';

export type BuilderErrorBannerKey =
  | 'builder.errors.workflowProblems'
  | 'builder.errors.badRequest'
  | 'builder.errors.unauthorized'
  | 'builder.errors.network'
  | 'builder.errors.unknown';

export type ParsedWorkflowErrors = {
  byNode: Record<string, string[]>;
  byField: Record<string, Record<string, string[]>>;
  global: string[];
  bannerKey: BuilderErrorBannerKey | null;
};

export function emptyWorkflowErrors(): ParsedWorkflowErrors {
  return {
    byNode: {},
    byField: {},
    global: [],
    bannerKey: null,
  };
}

/**
 * Extracts field names mentioned in a server validation message, such as:
 * - `"message": node "start" does not declare output "nope"`
 * - `invalid node configuration: "repo" is required; "issue_number" must be a number`
 * - `"right" is required`
 */
export function extractReferencedFields(message: string): string[] {
  const fields = new Set<string>();
  const refPrefix = /^"([a-z][a-z0-9_]*)":/;
  const prefixMatch = message.match(refPrefix);
  if (prefixMatch) {
    fields.add(prefixMatch[1]);
  }

  const schemaPattern =
    /"([a-z][a-z0-9_]*)"\s+(?:is required|must be a|is not a parameter|has a malformed)/g;
  for (const match of message.matchAll(schemaPattern)) {
    fields.add(match[1]);
  }

  return [...fields];
}

/**
 * Maps an API failure (including HTTP 422 `invalid_workflow` problems) to
 * step-level, field-level and workflow-level errors.
 */
export function extractWorkflowProblems(
  error: unknown,
  knownNodeIds: ReadonlySet<string>,
): ParsedWorkflowErrors {
  const result = emptyWorkflowErrors();
  if (!error) return result;

  if (!(error instanceof ApiError)) {
    result.bannerKey =
      error instanceof TypeError
        ? 'builder.errors.network'
        : 'builder.errors.unknown';
    return result;
  }

  if (error.status === 422) {
    result.bannerKey = 'builder.errors.workflowProblems';
    const body = error.body as Partial<WorkflowValidationError> | undefined;
    const problems = Array.isArray(body?.problems) ? body.problems : [];

    for (const problem of problems) {
      if (!problem || typeof problem.message !== 'string') continue;
      const nodeId = problem.node;
      if (nodeId && knownNodeIds.has(nodeId)) {
        result.byNode[nodeId] = [
          ...(result.byNode[nodeId] ?? []),
          problem.message,
        ];
        for (const fieldKey of extractReferencedFields(problem.message)) {
          const nodeFields = result.byField[nodeId] ?? {};
          nodeFields[fieldKey] = [
            ...(nodeFields[fieldKey] ?? []),
            problem.message,
          ];
          result.byField[nodeId] = nodeFields;
        }
      } else {
        result.global.push(
          nodeId ? `${nodeId}: ${problem.message}` : problem.message,
        );
      }
    }
    return result;
  }

  if (error.status === 400) {
    result.bannerKey = 'builder.errors.badRequest';
  } else if (error.status === 401) {
    result.bannerKey = 'builder.errors.unauthorized';
  } else {
    result.bannerKey = 'builder.errors.unknown';
  }
  return result;
}
