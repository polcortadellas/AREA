import { useQueryClient } from '@tanstack/react-query';

import {
  getGetWorkflowQueryKey,
  getListWorkflowsQueryKey,
  useDeleteWorkflow,
  useListWorkflows,
  useUpdateWorkflow,
} from '@/shared/api/generated/workflows/workflows';
import {
  getListRunsQueryKey,
  useTriggerWorkflow,
} from '@/shared/api/generated/runs/runs';
import type { Workflow, WorkflowSummary } from '@/shared/api/generated/model';

export function useWorkflowsList() {
  return useListWorkflows();
}

export function useToggleWorkflowStatus() {
  const queryClient = useQueryClient();

  return useUpdateWorkflow({
    mutation: {
      onMutate: async ({ workflowId, data }) => {
        const listKey = getListWorkflowsQueryKey();
        const detailKey = getGetWorkflowQueryKey(workflowId);
        await queryClient.cancelQueries({ queryKey: listKey });
        await queryClient.cancelQueries({ queryKey: detailKey });

        const previousList =
          queryClient.getQueryData<WorkflowSummary[]>(listKey);
        const previousDetail = queryClient.getQueryData<Workflow>(detailKey);

        if (previousList && data.enabled !== undefined) {
          queryClient.setQueryData<WorkflowSummary[]>(
            listKey,
            previousList.map(wf =>
              wf.id === workflowId ? { ...wf, enabled: data.enabled! } : wf,
            ),
          );
        }
        if (previousDetail && data.enabled !== undefined) {
          queryClient.setQueryData<Workflow>(detailKey, {
            ...previousDetail,
            enabled: data.enabled,
          });
        }

        return { previousList, previousDetail, workflowId };
      },
      onError: (_err, _vars, context) => {
        if (!context) return;
        if (context.previousList) {
          queryClient.setQueryData(
            getListWorkflowsQueryKey(),
            context.previousList,
          );
        }
        if (context.previousDetail) {
          queryClient.setQueryData(
            getGetWorkflowQueryKey(context.workflowId),
            context.previousDetail,
          );
        }
      },
      onSuccess: (updated, { workflowId }) => {
        queryClient.setQueryData<Workflow>(
          getGetWorkflowQueryKey(workflowId),
          updated,
        );
        queryClient.setQueryData<WorkflowSummary[]>(
          getListWorkflowsQueryKey(),
          current =>
            current?.map(wf =>
              wf.id === workflowId
                ? {
                    ...wf,
                    name: updated.name,
                    enabled: updated.enabled,
                    updated_at: updated.updated_at,
                  }
                : wf,
            ),
        );
      },
      onSettled: (_data, _error, { workflowId }) => {
        void queryClient.invalidateQueries({
          queryKey: getListWorkflowsQueryKey(),
        });
        void queryClient.invalidateQueries({
          queryKey: getGetWorkflowQueryKey(workflowId),
        });
      },
    },
  });
}

export function useDeleteWorkflowAction() {
  const queryClient = useQueryClient();

  return useDeleteWorkflow({
    mutation: {
      onSuccess: (_data, { workflowId }) => {
        queryClient.setQueryData<WorkflowSummary[]>(
          getListWorkflowsQueryKey(),
          current => current?.filter(wf => wf.id !== workflowId),
        );
        queryClient.removeQueries({
          queryKey: getGetWorkflowQueryKey(workflowId),
        });
        queryClient.removeQueries({
          queryKey: getListRunsQueryKey(workflowId),
        });
        void queryClient.invalidateQueries({
          queryKey: getListWorkflowsQueryKey(),
        });
      },
    },
  });
}

export function useTriggerWorkflowAction() {
  const queryClient = useQueryClient();

  return useTriggerWorkflow({
    mutation: {
      onSuccess: (_data, { workflowId }) => {
        void queryClient.invalidateQueries({
          queryKey: getListRunsQueryKey(workflowId),
        });
      },
    },
  });
}
