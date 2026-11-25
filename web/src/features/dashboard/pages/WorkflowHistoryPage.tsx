import { useMemo, useState } from 'react';
import { ArrowLeft, Pencil, Play, RefreshCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useParams } from 'react-router';

import { cn } from '@/shared/lib/utils';
import { Button } from '@/shared/ui/button';
import { RunListItem } from '../components/RunListItem';
import { RunStepTimeline } from '../components/RunStepTimeline';
import { WorkflowStatusBadge } from '../components/StatusBadge';
import {
  useToggleWorkflowStatus,
  useTriggerWorkflowAction,
} from '../hooks/useWorkflowDashboard';
import {
  useRunDetail,
  useWorkflowContext,
  useWorkflowRuns,
} from '../hooks/useWorkflowHistory';
import { dashboardErrorKey } from '../lib/errors';

const LIMIT_OPTIONS = [20, 50, 100] as const;

export function WorkflowHistoryPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { workflowId = '', runId } = useParams();

  const [limit, setLimit] = useState<number>(20);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const { workflow, summary } = useWorkflowContext(workflowId);
  const runsQuery = useWorkflowRuns(workflowId, limit);
  const runs = useMemo(() => runsQuery.data ?? [], [runsQuery.data]);

  const selectedRunId = runId ?? runs[0]?.id;
  const runDetailQuery = useRunDetail(selectedRunId);

  const toggleMutation = useToggleWorkflowStatus();
  const triggerMutation = useTriggerWorkflowAction();

  const workflowName = summary?.name ?? t('dashboard.history.fallbackName');
  const isPaused = summary ? !summary.enabled : false;

  const handleToggle = () => {
    if (!summary) return;
    setFeedback(null);
    setActionError(null);
    const nextEnabled = !summary.enabled;

    toggleMutation.mutate(
      { workflowId: summary.id, data: { enabled: nextEnabled } },
      {
        onSuccess: () => {
          setFeedback(
            t(
              nextEnabled
                ? 'dashboard.feedback.resumed'
                : 'dashboard.feedback.paused',
              { name: summary.name },
            ),
          );
        },
        onError: err => {
          setActionError(t(dashboardErrorKey(err)));
        },
      },
    );
  };

  const handleTrigger = () => {
    if (!workflowId || isPaused) return;
    setFeedback(null);
    setActionError(null);

    triggerMutation.mutate(
      { workflowId },
      {
        onSuccess: accepted => {
          setFeedback(
            t('dashboard.feedback.triggered', { name: workflowName }),
          );
          if (accepted?.id) {
            void navigate(`/workflows/${workflowId}/runs/${accepted.id}`);
          }
        },
        onError: err => {
          setActionError(t(dashboardErrorKey(err)));
        },
      },
    );
  };

  const handleRefresh = () => {
    void runsQuery.refetch();
    if (selectedRunId) {
      void runDetailQuery.refetch();
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <Link
          to="/"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground hover:underline"
        >
          <ArrowLeft className="size-4" aria-hidden="true" />
          {t('dashboard.history.backToDashboard')}
        </Link>

        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2.5">
              <h1 className="text-2xl font-semibold">{workflowName}</h1>
              {summary && <WorkflowStatusBadge enabled={summary.enabled} />}
            </div>
            <p className="text-sm text-muted-foreground">
              {t('dashboard.history.subtitle')}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2.5">
            {summary && (
              <div className="flex items-center gap-2 rounded-md border bg-card px-3 py-1.5">
                <span className="text-xs font-medium text-muted-foreground">
                  {toggleMutation.isPending
                    ? summary.enabled
                      ? t('dashboard.card.pausing')
                      : t('dashboard.card.resuming')
                    : summary.enabled
                      ? t('dashboard.card.pause')
                      : t('dashboard.card.resume')}
                </span>
                <button
                  type="button"
                  role="switch"
                  aria-checked={summary.enabled}
                  aria-label={t('dashboard.card.toggleLabel', {
                    name: summary.name,
                  })}
                  disabled={toggleMutation.isPending}
                  onClick={handleToggle}
                  className={cn(
                    'relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50',
                    summary.enabled
                      ? 'bg-emerald-600'
                      : 'bg-muted-foreground/35',
                  )}
                >
                  <span
                    aria-hidden="true"
                    className={cn(
                      'pointer-events-none block size-4 rounded-full bg-white shadow-xs transition-transform',
                      summary.enabled ? 'translate-x-4' : 'translate-x-0',
                    )}
                  />
                </button>
              </div>
            )}

            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isPaused || triggerMutation.isPending}
              title={
                isPaused ? t('dashboard.card.runNowDisabledPaused') : undefined
              }
              onClick={handleTrigger}
            >
              <Play aria-hidden="true" />
              {triggerMutation.isPending
                ? t('dashboard.card.running')
                : t('dashboard.card.runNow')}
            </Button>

            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={runsQuery.isFetching || runDetailQuery.isFetching}
              onClick={handleRefresh}
            >
              <RefreshCw
                aria-hidden="true"
                className={
                  runsQuery.isFetching || runDetailQuery.isFetching
                    ? 'animate-spin'
                    : undefined
                }
              />
              {t('dashboard.refresh')}
            </Button>

            <Button variant="ghost" size="sm" asChild>
              <Link to={`/workflows/${workflowId}/edit`}>
                <Pencil aria-hidden="true" />
                {t('dashboard.card.edit')}
              </Link>
            </Button>
          </div>
        </div>
      </div>

      {isPaused && (
        <p className="rounded-md border border-amber-200 bg-amber-50/70 px-4 py-2.5 text-sm text-amber-900 dark:border-amber-800/60 dark:bg-amber-950/40 dark:text-amber-200">
          {t('dashboard.card.pausedNotice')}
        </p>
      )}

      {feedback && (
        <div
          role="status"
          aria-live="polite"
          className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-sm text-emerald-900 dark:border-emerald-800/60 dark:bg-emerald-950/50 dark:text-emerald-200"
        >
          {feedback}
        </div>
      )}

      {actionError && (
        <div
          role="alert"
          className="rounded-md border border-destructive/40 bg-destructive/10 px-4 py-2.5 text-sm text-destructive"
        >
          {actionError}
        </div>
      )}

      {runsQuery.isLoading ? (
        <p
          role="status"
          className="rounded-lg border p-8 text-center text-sm text-muted-foreground"
        >
          {t('dashboard.history.loadingRuns')}
        </p>
      ) : runsQuery.isError ? (
        <div
          role="alert"
          className="space-y-3 rounded-lg border border-destructive/40 bg-destructive/10 p-6 text-sm text-destructive"
        >
          <p className="font-medium">{t('dashboard.history.loadRunsError')}</p>
          <p>{t(dashboardErrorKey(runsQuery.error))}</p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void runsQuery.refetch()}
          >
            {t('dashboard.errors.retry')}
          </Button>
        </div>
      ) : runs.length === 0 ? (
        <div className="space-y-3 rounded-lg border border-dashed p-10 text-center">
          <p className="text-base font-medium">
            {t('dashboard.history.emptyRunsTitle')}
          </p>
          <p className="text-sm text-muted-foreground">
            {t('dashboard.history.emptyRunsDescription')}
          </p>
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-[20rem_1fr]">
          <section className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-baseline gap-2">
                <h2 className="text-base font-semibold">
                  {t('dashboard.history.runsHeading')}
                </h2>
                <span className="text-xs text-muted-foreground">
                  {t('dashboard.history.runsCount', { count: runs.length })}
                </span>
              </div>

              <label className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <span className="sr-only">
                  {t('dashboard.history.limitLabel')}
                </span>
                <select
                  aria-label={t('dashboard.history.limitLabel')}
                  value={limit}
                  onChange={event => setLimit(Number(event.target.value))}
                  className="rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
                >
                  {LIMIT_OPTIONS.map(option => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <ul className="space-y-2">
              {runs.map(run => (
                <RunListItem
                  key={run.id}
                  workflowId={workflowId}
                  run={run}
                  isSelected={run.id === selectedRunId}
                />
              ))}
            </ul>
          </section>

          <section aria-label={t('dashboard.history.runHeading')}>
            {!selectedRunId ? (
              <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                {t('dashboard.history.selectRunPrompt')}
              </div>
            ) : runDetailQuery.isLoading ? (
              <p
                role="status"
                className="rounded-lg border p-8 text-center text-sm text-muted-foreground"
              >
                {t('dashboard.history.loadingRunDetail')}
              </p>
            ) : runDetailQuery.isError ? (
              <div
                role="alert"
                className="space-y-3 rounded-lg border border-destructive/40 bg-destructive/10 p-6 text-sm text-destructive"
              >
                <p className="font-medium">
                  {t('dashboard.history.loadRunDetailError')}
                </p>
                <p>{t(dashboardErrorKey(runDetailQuery.error))}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void runDetailQuery.refetch()}
                >
                  {t('dashboard.errors.retry')}
                </Button>
              </div>
            ) : runDetailQuery.data ? (
              <RunStepTimeline run={runDetailQuery.data} workflow={workflow} />
            ) : null}
          </section>
        </div>
      )}
    </div>
  );
}
