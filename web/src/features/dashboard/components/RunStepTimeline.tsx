import { useTranslation } from 'react-i18next';

import type {
  RunDetail,
  Workflow,
  WorkflowNode,
} from '@/shared/api/generated/model';
import {
  formatDateTime,
  formatDuration,
  formatJson,
  hasRecordEntries,
  orderRunSteps,
} from '../lib/format';
import { RunStatusBadge, StepStatusBadge } from './StatusBadge';

interface RunStepTimelineProps {
  run: RunDetail;
  workflow?: Workflow | null;
}

export function RunStepTimeline({ run, workflow }: RunStepTimelineProps) {
  const { t, i18n } = useTranslation();

  const nodeIndex = new Map<string, WorkflowNode>(
    (workflow?.nodes ?? []).map(node => [node.id, node]),
  );
  const orderedSteps = orderRunSteps(run.steps, workflow);

  const created = formatDateTime(run.created_at, i18n.language);
  const started = formatDateTime(run.started_at, i18n.language);
  const finished = formatDateTime(run.finished_at, i18n.language);
  const duration = formatDuration(run.started_at, run.finished_at);
  const na = t('dashboard.history.notAvailable');

  return (
    <div className="space-y-6">
      <section className="rounded-lg border bg-card p-5 shadow-xs">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">
            {t('dashboard.history.runHeading')}
          </h2>
          <RunStatusBadge status={run.status} />
        </div>

        <dl className="mt-4 grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
          <div className="sm:col-span-2 lg:col-span-3">
            <dt className="text-xs font-medium text-muted-foreground">
              {t('dashboard.history.runId')}
            </dt>
            <dd className="mt-0.5 font-mono text-xs break-all">{run.id}</dd>
          </div>
          <div>
            <dt className="text-xs font-medium text-muted-foreground">
              {t('dashboard.history.createdAt')}
            </dt>
            <dd className="mt-0.5">{created ?? na}</dd>
          </div>
          <div>
            <dt className="text-xs font-medium text-muted-foreground">
              {t('dashboard.history.startedAt')}
            </dt>
            <dd className="mt-0.5">{started ?? na}</dd>
          </div>
          <div>
            <dt className="text-xs font-medium text-muted-foreground">
              {t('dashboard.history.finishedAt')}
            </dt>
            <dd className="mt-0.5">{finished ?? na}</dd>
          </div>
          <div>
            <dt className="text-xs font-medium text-muted-foreground">
              {t('dashboard.history.duration')}
            </dt>
            <dd className="mt-0.5">{duration ?? na}</dd>
          </div>
        </dl>

        {run.error && (
          <div
            role="alert"
            className="mt-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
          >
            <p className="font-medium">{t('dashboard.history.runError')}</p>
            <p className="mt-1 font-mono text-xs break-words">{run.error}</p>
          </div>
        )}

        <div className="mt-4 border-t pt-4">
          <h3 className="text-sm font-medium">
            {t('dashboard.history.triggerPayload')}
          </h3>
          {hasRecordEntries(run.trigger_payload) ? (
            <pre className="mt-2 overflow-x-auto rounded-md border bg-muted/50 p-3 font-mono text-xs">
              <code>{formatJson(run.trigger_payload)}</code>
            </pre>
          ) : (
            <p className="mt-1.5 text-xs text-muted-foreground">
              {t('dashboard.history.emptyPayload')}
            </p>
          )}
        </div>
      </section>

      <section className="space-y-3">
        <div className="flex items-baseline justify-between gap-2">
          <h3 className="text-base font-semibold">
            {t('dashboard.history.stepsHeading')}
          </h3>
          <span className="text-xs text-muted-foreground">
            {t('dashboard.history.stepsCount', { count: orderedSteps.length })}
          </span>
        </div>

        {orderedSteps.length === 0 ? (
          <div className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
            {t('dashboard.history.emptySteps')}
          </div>
        ) : (
          <ol className="space-y-3">
            {orderedSteps.map((step, index) => {
              const nodeMeta = nodeIndex.get(step.node_id);
              const stepStarted = formatDateTime(
                step.started_at,
                i18n.language,
              );
              const stepFinished = formatDateTime(
                step.finished_at,
                i18n.language,
              );
              const stepDuration = formatDuration(
                step.started_at,
                step.finished_at,
              );
              const hasOutput =
                step.output !== null && step.output !== undefined;

              return (
                <li
                  key={step.node_id}
                  className="rounded-lg border bg-card p-4 shadow-xs"
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0 space-y-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="inline-flex rounded bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground">
                          {t('dashboard.history.stepNumber', {
                            index: index + 1,
                          })}
                        </span>
                        {nodeMeta && (
                          <span className="inline-flex rounded border px-2 py-0.5 text-xs font-medium text-muted-foreground">
                            {t(`dashboard.history.nodeKind.${nodeMeta.kind}`)}
                          </span>
                        )}
                        <code className="text-sm font-semibold">
                          {step.node_id}
                        </code>
                        {nodeMeta && nodeMeta.type !== step.node_id && (
                          <code className="text-xs text-muted-foreground">
                            ({nodeMeta.type})
                          </code>
                        )}
                      </div>

                      <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                        <span>
                          {step.attempt > 0
                            ? t('dashboard.history.attempt', {
                                count: step.attempt,
                              })
                            : t('dashboard.history.notExecuted')}
                        </span>
                        {stepStarted && (
                          <span>
                            {t('dashboard.history.startedAt')}: {stepStarted}
                          </span>
                        )}
                        {stepFinished && (
                          <span>
                            {t('dashboard.history.finishedAt')}: {stepFinished}
                          </span>
                        )}
                        {stepDuration && (
                          <span>
                            {t('dashboard.history.duration')}: {stepDuration}
                          </span>
                        )}
                      </div>
                    </div>

                    <StepStatusBadge status={step.status} />
                  </div>

                  {step.status === 'SKIPPED' && (
                    <p className="mt-3 rounded-md bg-muted/60 px-3 py-2 text-xs text-muted-foreground">
                      {t('dashboard.history.skippedNotice')}
                    </p>
                  )}

                  {step.error && (
                    <div
                      role="alert"
                      className="mt-3 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive"
                    >
                      <p className="font-medium">
                        {t('dashboard.history.stepError')}
                      </p>
                      <p className="mt-1 font-mono break-words">{step.error}</p>
                    </div>
                  )}

                  <div className="mt-3 border-t pt-3">
                    <p className="text-xs font-medium text-muted-foreground">
                      {t('dashboard.history.output')}
                    </p>
                    {hasOutput ? (
                      <pre className="mt-1.5 overflow-x-auto rounded-md border bg-muted/50 p-3 font-mono text-xs">
                        <code>{formatJson(step.output)}</code>
                      </pre>
                    ) : (
                      <p className="mt-1 text-xs text-muted-foreground">
                        {t('dashboard.history.noOutput')}
                      </p>
                    )}
                  </div>
                </li>
              );
            })}
          </ol>
        )}
      </section>
    </div>
  );
}
