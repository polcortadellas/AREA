import { useGetRun, useListRuns } from '@/shared/api/generated/runs/runs';
import {
  useGetWorkflow,
  useListWorkflows,
} from '@/shared/api/generated/workflows/workflows';
import type { RunStatus, WorkflowSummary } from '@/shared/api/generated/model';

const ACTIVE_RUN_STATUSES = new Set<RunStatus>(['PENDING', 'RUNNING']);
const POLL_INTERVAL_MS = 3000;

export function useWorkflowRuns(workflowId: string, limit: number) {
  return useListRuns(
    workflowId,
    { limit },
    {
      query: {
        enabled: Boolean(workflowId),
        refetchInterval: query => {
          const runs = query.state.data;
          if (!runs) return false;
          return runs.some(r => ACTIVE_RUN_STATUSES.has(r.status))
            ? POLL_INTERVAL_MS
            : false;
        },
      },
    },
  );
}

export function useRunDetail(runId: string | undefined) {
  return useGetRun(runId ?? '', {
    query: {
      enabled: Boolean(runId),
      refetchInterval: query => {
        const run = query.state.data;
        if (!run) return false;
        return ACTIVE_RUN_STATUSES.has(run.status) ? POLL_INTERVAL_MS : false;
      },
    },
  });
}

/**
 * Resolves the workflow header and optional node metadata for the history view.
 * Falls back gracefully to the workflow summary list if GET /workflows/:id is unavailable.
 */
export function useWorkflowContext(workflowId: string) {
  const detailQuery = useGetWorkflow(workflowId, {
    query: {
      enabled: Boolean(workflowId),
      retry: false,
    },
  });
  const listQuery = useListWorkflows();

  const summary: WorkflowSummary | undefined =
    detailQuery.data ?? listQuery.data?.find(w => w.id === workflowId);

  return {
    workflow: detailQuery.data ?? null,
    summary: summary ?? null,
    isLoading: detailQuery.isLoading && listQuery.isLoading,
  };
}
