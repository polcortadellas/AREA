import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router';

import { ConditionStepCard } from '@/features/builder/components/ConditionStepCard';
import { ReactionStepCard } from '@/features/builder/components/ReactionStepCard';
import { TriggerStepCard } from '@/features/builder/components/TriggerStepCard';
import {
  buildUpstreamOutputs,
  defaultInputValues,
  findNodeByType,
  findServiceByName,
  reactionServices,
  triggerServices,
} from '@/features/builder/lib/catalogue';
import {
  emptyWorkflowErrors,
  extractWorkflowProblems,
  type ParsedWorkflowErrors,
} from '@/features/builder/lib/errors';
import {
  activeNodeIds,
  createDefaultReaction,
  createInitialDraft,
  fromWorkflow,
  nextReactionId,
  toWorkflowInput,
  type DraftReaction,
  type WorkflowDraft,
} from '@/features/builder/lib/workflowGraph';
import type { About, Workflow } from '@/shared/api/generated/model';
import { useGetAbout } from '@/shared/api/generated/system/system';
import {
  useCreateWorkflow,
  useGetWorkflow,
  useReplaceWorkflow,
} from '@/shared/api/generated/workflows/workflows';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';

type ReactionListKey =
  'reactions' | 'trueBranchReactions' | 'falseBranchReactions';

export function WorkflowBuilderPage() {
  const { t } = useTranslation();
  const { workflowId } = useParams<{ workflowId?: string }>();
  const aboutQuery = useGetAbout();
  const workflowQuery = useGetWorkflow(workflowId ?? '', {
    query: { enabled: Boolean(workflowId) },
  });

  if (aboutQuery.isPending || (workflowId && workflowQuery.isPending)) {
    return (
      <section aria-busy="true" className="py-8">
        <h1 className="text-2xl font-semibold">
          {workflowId ? t('builder.editTitle') : t('builder.title')}
        </h1>
        <p className="text-muted-foreground mt-2 text-sm">
          {t('builder.loading')}
        </p>
      </section>
    );
  }

  if (
    aboutQuery.isError ||
    !aboutQuery.data ||
    (workflowId && workflowQuery.isError)
  ) {
    return (
      <section className="py-8">
        <h1 className="text-2xl font-semibold">
          {workflowId ? t('builder.editTitle') : t('builder.title')}
        </h1>
        <p role="alert" className="text-destructive mt-2 text-sm">
          {t('builder.loadError')}
        </p>
      </section>
    );
  }

  const about = aboutQuery.data;
  if (
    triggerServices(about).length === 0 ||
    reactionServices(about).length === 0
  ) {
    return (
      <section className="py-8">
        <h1 className="text-2xl font-semibold">
          {workflowId ? t('builder.editTitle') : t('builder.title')}
        </h1>
        <p role="alert" className="text-muted-foreground mt-2 text-sm">
          {t('builder.emptyCatalogue')}
        </p>
      </section>
    );
  }

  return (
    <WorkflowBuilderForm
      key={workflowId ? `edit-${workflowId}` : 'create'}
      about={about}
      workflowId={workflowId}
      existingWorkflow={workflowQuery.data}
    />
  );
}

function WorkflowBuilderForm({
  about,
  workflowId,
  existingWorkflow,
}: {
  about: About;
  workflowId?: string;
  existingWorkflow?: Workflow;
}) {
  const { t } = useTranslation();
  const [draft, setDraft] = useState<WorkflowDraft>(() =>
    existingWorkflow
      ? fromWorkflow(existingWorkflow, about)
      : createInitialDraft(about),
  );
  const [nameError, setNameError] = useState(false);
  const [serverErrors, setServerErrors] =
    useState<ParsedWorkflowErrors>(emptyWorkflowErrors);
  const [saved, setSaved] = useState(false);

  const createMutation = useCreateWorkflow();
  const replaceMutation = useReplaceWorkflow();
  const isSubmitting = createMutation.isPending || replaceMutation.isPending;

  const updateDraft = (updater: (prev: WorkflowDraft) => WorkflowDraft) => {
    setSaved(false);
    setDraft(updater);
  };

  const handleTriggerServiceChange = (serviceName: string) => {
    const service = findServiceByName(about, serviceName);
    const firstAction = service?.actions[0];
    updateDraft(prev => ({
      ...prev,
      trigger: {
        ...prev.trigger,
        service: serviceName,
        type: firstAction?.type ?? '',
        config: defaultInputValues(firstAction?.inputs ?? []),
      },
    }));
  };

  const handleTriggerActionChange = (actionType: string) => {
    const action = findNodeByType(about, 'trigger', actionType);
    updateDraft(prev => ({
      ...prev,
      trigger: {
        ...prev.trigger,
        type: actionType,
        config: defaultInputValues(action?.inputs ?? []),
      },
    }));
  };

  const handleTriggerConfigChange = (key: string, value: string | boolean) => {
    updateDraft(prev => ({
      ...prev,
      trigger: {
        ...prev.trigger,
        config: { ...prev.trigger.config, [key]: value },
      },
    }));
  };

  const handleAddCondition = () => {
    updateDraft(prev => ({
      ...prev,
      condition: {
        id: 'condition',
        left: '',
        op: 'equal',
        right: '',
      },
      trueBranchReactions:
        prev.trueBranchReactions.length > 0
          ? prev.trueBranchReactions
          : [createDefaultReaction(about, 'true_1')],
      falseBranchReactions:
        prev.falseBranchReactions.length > 0
          ? prev.falseBranchReactions
          : [createDefaultReaction(about, 'false_1')],
    }));
  };

  const handleRemoveCondition = () => {
    updateDraft(prev => ({
      ...prev,
      condition: null,
    }));
  };

  const updateReactionList = (
    listKey: ReactionListKey,
    updater: (list: DraftReaction[]) => DraftReaction[],
  ) => {
    updateDraft(prev => ({
      ...prev,
      [listKey]: updater([...prev[listKey]]),
    }));
  };

  const prefixForList = (listKey: ReactionListKey) =>
    listKey === 'trueBranchReactions'
      ? 'true'
      : listKey === 'falseBranchReactions'
        ? 'false'
        : 'reaction';

  const handleAddReaction = (listKey: ReactionListKey) => {
    updateReactionList(listKey, list => [
      ...list,
      createDefaultReaction(
        about,
        nextReactionId(prefixForList(listKey), list),
      ),
    ]);
  };

  const handleRemoveReaction = (listKey: ReactionListKey, index: number) => {
    updateReactionList(listKey, list => list.filter((_, i) => i !== index));
  };

  const handleMoveReaction = (
    listKey: ReactionListKey,
    index: number,
    direction: -1 | 1,
  ) => {
    updateReactionList(listKey, list => {
      const target = index + direction;
      if (target < 0 || target >= list.length) return list;
      const copy = [...list];
      const [moved] = copy.splice(index, 1);
      copy.splice(target, 0, moved);
      return copy;
    });
  };

  const handleReactionServiceChange = (
    listKey: ReactionListKey,
    index: number,
    serviceName: string,
  ) => {
    const service = findServiceByName(about, serviceName);
    const firstReaction = service?.reactions[0];
    updateReactionList(listKey, list =>
      list.map((item, i) =>
        i === index
          ? {
              ...item,
              service: serviceName,
              type: firstReaction?.type ?? '',
              config: defaultInputValues(firstReaction?.inputs ?? []),
            }
          : item,
      ),
    );
  };

  const handleReactionTypeChange = (
    listKey: ReactionListKey,
    index: number,
    reactionType: string,
  ) => {
    const reactionDef = findNodeByType(about, 'reaction', reactionType);
    updateReactionList(listKey, list =>
      list.map((item, i) =>
        i === index
          ? {
              ...item,
              type: reactionType,
              config: defaultInputValues(reactionDef?.inputs ?? []),
            }
          : item,
      ),
    );
  };

  const handleReactionConfigChange = (
    listKey: ReactionListKey,
    index: number,
    key: string,
    value: string | boolean,
  ) => {
    updateReactionList(listKey, list =>
      list.map((item, i) =>
        i === index
          ? {
              ...item,
              config: { ...item.config, [key]: value },
            }
          : item,
      ),
    );
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaved(false);

    if (!draft.name.trim()) {
      setNameError(true);
      return;
    }
    setNameError(false);
    setServerErrors(emptyWorkflowErrors());

    const payload = toWorkflowInput(draft, about);
    try {
      if (workflowId) {
        await replaceMutation.mutateAsync({ workflowId, data: payload });
      } else {
        await createMutation.mutateAsync({ data: payload });
      }
      setSaved(true);
    } catch (error) {
      setServerErrors(extractWorkflowProblems(error, activeNodeIds(draft)));
    }
  };

  const renderReactionChain = (
    listKey: ReactionListKey,
    branchLabel?: string,
    allowEmpty = false,
  ) => {
    const list = draft[listKey];
    if (list.length === 0) {
      return (
        <p className="text-muted-foreground rounded-md border border-dashed p-4 text-sm">
          {t('builder.condition.emptyBranch')}
        </p>
      );
    }

    return (
      <div className="space-y-4">
        {list.map((reaction, index) => {
          const stepTitle = branchLabel
            ? t('builder.reactions.branchStepHeading', {
                branch: branchLabel,
                index: index + 1,
              })
            : t('builder.reactions.stepHeading', { index: index + 1 });
          const upstreamOutputs = buildUpstreamOutputs(
            about,
            draft.trigger.type,
            list.slice(0, index),
          );

          return (
            <ReactionStepCard
              key={reaction.id}
              about={about}
              reaction={reaction}
              stepTitle={stepTitle}
              upstreamOutputs={upstreamOutputs}
              problems={serverErrors.byNode[reaction.id]}
              fieldErrors={serverErrors.byField[reaction.id]}
              canMoveUp={index > 0}
              canMoveDown={index < list.length - 1}
              canRemove={allowEmpty || list.length > 1}
              onServiceChange={serviceName =>
                handleReactionServiceChange(listKey, index, serviceName)
              }
              onReactionChange={reactionType =>
                handleReactionTypeChange(listKey, index, reactionType)
              }
              onConfigChange={(key, value) =>
                handleReactionConfigChange(listKey, index, key, value)
              }
              onMoveUp={() => handleMoveReaction(listKey, index, -1)}
              onMoveDown={() => handleMoveReaction(listKey, index, 1)}
              onRemove={() => handleRemoveReaction(listKey, index)}
            />
          );
        })}
      </div>
    );
  };

  const showTopErrorBanner =
    serverErrors.global.length > 0 ||
    (serverErrors.bannerKey !== null &&
      serverErrors.bannerKey !== 'builder.errors.workflowProblems');

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">
          {workflowId ? t('builder.editTitle') : t('builder.title')}
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          {t('builder.subtitle')}
        </p>
      </div>

      <form onSubmit={handleSubmit} noValidate className="space-y-6">
        <div className="space-y-4 rounded-lg border p-5 shadow-xs">
          <div className="space-y-1.5">
            <Label htmlFor="workflow-name">{t('builder.nameLabel')}</Label>
            <Input
              id="workflow-name"
              type="text"
              placeholder={t('builder.namePlaceholder')}
              value={draft.name}
              onChange={e => {
                if (nameError) setNameError(false);
                updateDraft(prev => ({ ...prev, name: e.target.value }));
              }}
              aria-invalid={nameError}
              aria-describedby={nameError ? 'workflow-name-error' : undefined}
            />
            {nameError && (
              <p
                id="workflow-name-error"
                role="alert"
                className="text-destructive text-sm"
              >
                {t('builder.errors.nameRequired')}
              </p>
            )}
          </div>

          <div className="flex items-center gap-2">
            <input
              id="workflow-enabled"
              type="checkbox"
              checked={draft.enabled}
              onChange={e =>
                updateDraft(prev => ({ ...prev, enabled: e.target.checked }))
              }
              className="size-4 rounded border"
            />
            <Label htmlFor="workflow-enabled">
              {t('builder.enabledLabel')}
            </Label>
          </div>
        </div>

        <TriggerStepCard
          about={about}
          trigger={draft.trigger}
          problems={serverErrors.byNode[draft.trigger.id]}
          fieldErrors={serverErrors.byField[draft.trigger.id]}
          onServiceChange={handleTriggerServiceChange}
          onActionChange={handleTriggerActionChange}
          onConfigChange={handleTriggerConfigChange}
        />

        {!draft.condition ? (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 className="text-lg font-semibold">
                  {t('builder.reactions.heading')}
                </h2>
                <p className="text-muted-foreground text-sm">
                  {t('builder.reactions.description')}
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddCondition}
              >
                {t('builder.condition.add')}
              </Button>
            </div>

            {renderReactionChain('reactions')}

            <div>
              <Button
                type="button"
                variant="secondary"
                onClick={() => handleAddReaction('reactions')}
              >
                {t('builder.reactions.add')}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-6">
            <ConditionStepCard
              condition={draft.condition}
              upstreamOutputs={buildUpstreamOutputs(
                about,
                draft.trigger.type,
                [],
              )}
              problems={serverErrors.byNode[draft.condition.id]}
              fieldErrors={serverErrors.byField[draft.condition.id]}
              onChange={nextCondition =>
                updateDraft(prev => ({ ...prev, condition: nextCondition }))
              }
              onRemove={handleRemoveCondition}
            />

            <div className="grid gap-6 lg:grid-cols-2">
              <div
                data-testid="branch-true"
                className="space-y-4 rounded-lg border p-4"
              >
                <div>
                  <h2 className="text-lg font-semibold">
                    {t('builder.condition.trueBranch')}
                  </h2>
                  <p className="text-muted-foreground text-sm">
                    {t('builder.condition.trueBranchDescription')}
                  </p>
                </div>

                {renderReactionChain(
                  'trueBranchReactions',
                  t('builder.condition.trueBranch'),
                  true,
                )}

                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={() => handleAddReaction('trueBranchReactions')}
                >
                  {t('builder.condition.addTrueReaction')}
                </Button>
              </div>

              <div
                data-testid="branch-false"
                className="space-y-4 rounded-lg border p-4"
              >
                <div>
                  <h2 className="text-lg font-semibold">
                    {t('builder.condition.falseBranch')}
                  </h2>
                  <p className="text-muted-foreground text-sm">
                    {t('builder.condition.falseBranchDescription')}
                  </p>
                </div>

                {renderReactionChain(
                  'falseBranchReactions',
                  t('builder.condition.falseBranch'),
                  true,
                )}

                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={() => handleAddReaction('falseBranchReactions')}
                >
                  {t('builder.condition.addFalseReaction')}
                </Button>
              </div>
            </div>
          </div>
        )}

        {showTopErrorBanner && (
          <div
            role="alert"
            className="border-destructive/50 bg-destructive/10 text-destructive rounded-md border p-3 text-sm"
          >
            {serverErrors.bannerKey && (
              <p className="font-medium">{t(serverErrors.bannerKey)}</p>
            )}
            {serverErrors.global.length > 0 && (
              <ul className="mt-1 list-inside list-disc space-y-0.5">
                {serverErrors.global.map((msg, idx) => (
                  <li key={`global-error-${idx}`}>{msg}</li>
                ))}
              </ul>
            )}
          </div>
        )}

        {saved && (
          <p
            role="status"
            className="rounded-md border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm font-medium text-emerald-700 dark:text-emerald-300"
          >
            {t('builder.actions.saved')}
          </p>
        )}

        <div className="flex items-center justify-end gap-3">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting
              ? t('builder.actions.saving')
              : t('builder.actions.save')}
          </Button>
        </div>
      </form>
    </section>
  );
}
