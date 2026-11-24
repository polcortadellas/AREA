import { useMemo, useState } from 'react';
import { Plus, RefreshCw, Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';

import type { WorkflowSummary } from '@/shared/api/generated/model';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { WorkflowCard } from '../components/WorkflowCard';
import {
  useDeleteWorkflowAction,
  useToggleWorkflowStatus,
  useTriggerWorkflowAction,
  useWorkflowsList,
} from '../hooks/useWorkflowDashboard';
import { dashboardErrorKey } from '../lib/errors';

type StatusFilter = 'all' | 'active' | 'paused';

export function DashboardPage() {
  const { t } = useTranslation();
  const [filter, setFilter] = useState<StatusFilter>('all');
  const [search, setSearch] = useState('');
  const [feedback, setFeedback] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [triggeringId, setTriggeringId] = useState<string | null>(null);

  const workflowsQuery = useWorkflowsList();
  const toggleMutation = useToggleWorkflowStatus();
  const deleteMutation = useDeleteWorkflowAction();
  const triggerMutation = useTriggerWorkflowAction();

  const workflows = useMemo(
    () => workflowsQuery.data ?? [],
    [workflowsQuery.data],
  );

  const stats = useMemo(() => {
    const active = workflows.filter(wf => wf.enabled).length;
    return {
      total: workflows.length,
      active,
      paused: workflows.length - active,
    };
  }, [workflows]);

  const filteredWorkflows = useMemo(() => {
    const query = search.trim().toLowerCase();
    return workflows.filter(wf => {
      if (filter === 'active' && !wf.enabled) return false;
      if (filter === 'paused' && wf.enabled) return false;
      if (query && !wf.name.toLowerCase().includes(query)) return false;
      return true;
    });
  }, [workflows, filter, search]);

  const handleToggle = (workflow: WorkflowSummary) => {
    setFeedback(null);
    setActionError(null);
    setTogglingId(workflow.id);
    const nextEnabled = !workflow.enabled;

    toggleMutation.mutate(
      { workflowId: workflow.id, data: { enabled: nextEnabled } },
      {
        onSuccess: () => {
          setFeedback(
            t(
              nextEnabled
                ? 'dashboard.feedback.resumed'
                : 'dashboard.feedback.paused',
              { name: workflow.name },
            ),
          );
        },
        onError: err => {
          setActionError(t(dashboardErrorKey(err)));
        },
        onSettled: () => {
          setTogglingId(current => (current === workflow.id ? null : current));
        },
      },
    );
  };

  const handleDelete = (workflow: WorkflowSummary) => {
    setFeedback(null);
    setActionError(null);
    setDeletingId(workflow.id);

    deleteMutation.mutate(
      { workflowId: workflow.id },
      {
        onSuccess: () => {
          setFeedback(t('dashboard.feedback.deleted', { name: workflow.name }));
        },
        onError: err => {
          setActionError(t(dashboardErrorKey(err)));
        },
        onSettled: () => {
          setDeletingId(current => (current === workflow.id ? null : current));
        },
      },
    );
  };

  const handleTrigger = (workflow: WorkflowSummary) => {
    if (!workflow.enabled) return;
    setFeedback(null);
    setActionError(null);
    setTriggeringId(workflow.id);

    triggerMutation.mutate(
      { workflowId: workflow.id },
      {
        onSuccess: () => {
          setFeedback(
            t('dashboard.feedback.triggered', { name: workflow.name }),
          );
        },
        onError: err => {
          setActionError(t(dashboardErrorKey(err)));
        },
        onSettled: () => {
          setTriggeringId(current =>
            current === workflow.id ? null : current,
          );
        },
      },
    );
  };

  const hasFilters = filter !== 'all' || search.trim() !== '';

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-2xl font-semibold">{t('dashboard.title')}</h1>
          <p className="text-sm text-muted-foreground">
            {t('dashboard.subtitle')}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={workflowsQuery.isFetching}
            onClick={() => void workflowsQuery.refetch()}
          >
            <RefreshCw
              aria-hidden="true"
              className={workflowsQuery.isFetching ? 'animate-spin' : undefined}
            />
            {t('dashboard.refresh')}
          </Button>
          <Button asChild size="sm">
            <Link to="/workflows/new">
              <Plus aria-hidden="true" />
              {t('dashboard.newWorkflow')}
            </Link>
          </Button>
        </div>
      </div>

      {!workflowsQuery.isLoading && !workflowsQuery.isError && (
        <section aria-label={t('dashboard.stats.summary')}>
          <dl className="grid grid-cols-3 gap-3">
            <div className="rounded-lg border bg-card p-3.5 shadow-xs">
              <dt className="text-xs font-medium text-muted-foreground">
                {t('dashboard.stats.total')}
              </dt>
              <dd className="mt-1 text-2xl font-semibold tabular-nums">
                {stats.total}
              </dd>
            </div>
            <div className="rounded-lg border bg-card p-3.5 shadow-xs">
              <dt className="text-xs font-medium text-muted-foreground">
                {t('dashboard.stats.active')}
              </dt>
              <dd className="mt-1 text-2xl font-semibold text-emerald-700 tabular-nums dark:text-emerald-400">
                {stats.active}
              </dd>
            </div>
            <div className="rounded-lg border bg-card p-3.5 shadow-xs">
              <dt className="text-xs font-medium text-muted-foreground">
                {t('dashboard.stats.paused')}
              </dt>
              <dd className="mt-1 text-2xl font-semibold text-amber-800 tabular-nums dark:text-amber-400">
                {stats.paused}
              </dd>
            </div>
          </dl>
        </section>
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

      {!workflowsQuery.isLoading &&
        !workflowsQuery.isError &&
        workflows.length > 0 && (
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="relative min-w-[14rem] flex-1 sm:max-w-sm">
              <Label htmlFor="workflow-search" className="sr-only">
                {t('dashboard.filters.searchLabel')}
              </Label>
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                id="workflow-search"
                type="search"
                value={search}
                onChange={event => setSearch(event.target.value)}
                placeholder={t('dashboard.filters.searchPlaceholder')}
                className="pl-9"
              />
            </div>

            <div
              role="group"
              aria-label={t('dashboard.filters.label')}
              className="flex items-center gap-1.5"
            >
              {(['all', 'active', 'paused'] as const).map(option => (
                <Button
                  key={option}
                  type="button"
                  size="sm"
                  variant={filter === option ? 'default' : 'outline'}
                  aria-pressed={filter === option}
                  onClick={() => setFilter(option)}
                >
                  {t(`dashboard.filters.${option}`)}
                </Button>
              ))}
            </div>
          </div>
        )}

      {workflowsQuery.isLoading ? (
        <p
          role="status"
          className="rounded-lg border p-8 text-center text-sm text-muted-foreground"
        >
          {t('dashboard.loading')}
        </p>
      ) : workflowsQuery.isError ? (
        <div
          role="alert"
          className="space-y-3 rounded-lg border border-destructive/40 bg-destructive/10 p-6 text-sm text-destructive"
        >
          <p className="font-medium">{t('dashboard.errors.loadFailed')}</p>
          <p>{t(dashboardErrorKey(workflowsQuery.error))}</p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void workflowsQuery.refetch()}
          >
            {t('dashboard.errors.retry')}
          </Button>
        </div>
      ) : workflows.length === 0 ? (
        <div className="space-y-3 rounded-lg border border-dashed p-10 text-center">
          <p className="text-base font-medium">{t('dashboard.empty.title')}</p>
          <p className="text-sm text-muted-foreground">
            {t('dashboard.empty.description')}
          </p>
          <Button asChild size="sm">
            <Link to="/workflows/new">
              <Plus aria-hidden="true" />
              {t('dashboard.newWorkflow')}
            </Link>
          </Button>
        </div>
      ) : filteredWorkflows.length === 0 ? (
        <div className="space-y-3 rounded-lg border border-dashed p-10 text-center">
          <p className="text-base font-medium">
            {t('dashboard.empty.filteredTitle')}
          </p>
          <p className="text-sm text-muted-foreground">
            {t('dashboard.empty.filteredDescription')}
          </p>
          {hasFilters && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setFilter('all');
                setSearch('');
              }}
            >
              {t('dashboard.filters.clear')}
            </Button>
          )}
        </div>
      ) : (
        <ul className="space-y-4">
          {filteredWorkflows.map(workflow => (
            <li key={workflow.id}>
              <WorkflowCard
                workflow={workflow}
                isToggling={togglingId === workflow.id}
                isDeleting={deletingId === workflow.id}
                isTriggering={triggeringId === workflow.id}
                onToggle={handleToggle}
                onDelete={handleDelete}
                onTrigger={handleTrigger}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
