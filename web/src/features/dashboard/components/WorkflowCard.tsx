import { useState } from 'react';
import { History, Pencil, Play, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';

import type { WorkflowSummary } from '@/shared/api/generated/model';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/shared/ui/button';
import { formatDateTime } from '../lib/format';
import { WorkflowStatusBadge } from './StatusBadge';

interface WorkflowCardProps {
  workflow: WorkflowSummary;
  isToggling: boolean;
  isDeleting: boolean;
  isTriggering: boolean;
  onToggle: (workflow: WorkflowSummary) => void;
  onDelete: (workflow: WorkflowSummary) => void;
  onTrigger: (workflow: WorkflowSummary) => void;
}

export function WorkflowCard({
  workflow,
  isToggling,
  isDeleting,
  isTriggering,
  onToggle,
  onDelete,
  onTrigger,
}: WorkflowCardProps) {
  const { t, i18n } = useTranslation();
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const updatedLabel = formatDateTime(workflow.updated_at, i18n.language);
  const createdLabel = formatDateTime(workflow.created_at, i18n.language);

  return (
    <article className="rounded-lg border bg-card p-5 shadow-xs transition-colors">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0 flex-1 space-y-1.5">
          <div className="flex flex-wrap items-center gap-2.5">
            <h2 className="truncate text-lg font-semibold">
              <Link
                to={`/workflows/${workflow.id}/runs`}
                className="hover:underline focus-visible:rounded-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                {workflow.name}
              </Link>
            </h2>
            <WorkflowStatusBadge enabled={workflow.enabled} />
          </div>

          <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
            {updatedLabel && (
              <span>
                {t('dashboard.card.updatedAt', { date: updatedLabel })}
              </span>
            )}
            {createdLabel && (
              <span>
                {t('dashboard.card.createdAt', { date: createdLabel })}
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2.5">
          <span className="text-xs font-medium text-muted-foreground">
            {isToggling
              ? workflow.enabled
                ? t('dashboard.card.pausing')
                : t('dashboard.card.resuming')
              : workflow.enabled
                ? t('dashboard.card.pause')
                : t('dashboard.card.resume')}
          </span>
          <button
            type="button"
            role="switch"
            aria-checked={workflow.enabled}
            aria-label={t('dashboard.card.toggleLabel', {
              name: workflow.name,
            })}
            disabled={isToggling || isDeleting}
            onClick={() => onToggle(workflow)}
            className={cn(
              'relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50',
              workflow.enabled ? 'bg-emerald-600' : 'bg-muted-foreground/35',
            )}
          >
            <span
              aria-hidden="true"
              className={cn(
                'pointer-events-none block size-5 rounded-full bg-white shadow-xs transition-transform',
                workflow.enabled ? 'translate-x-5' : 'translate-x-0',
              )}
            />
          </button>
        </div>
      </div>

      {!workflow.enabled && (
        <p className="mt-3 rounded-md border border-amber-200 bg-amber-50/70 px-3 py-2 text-xs text-amber-900 dark:border-amber-800/60 dark:bg-amber-950/40 dark:text-amber-200">
          {t('dashboard.card.pausedNotice')}
        </p>
      )}

      <div className="mt-4 flex flex-wrap items-center justify-between gap-2 border-t pt-3.5">
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link to={`/workflows/${workflow.id}/runs`}>
              <History aria-hidden="true" />
              {t('dashboard.card.history')}
            </Link>
          </Button>

          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!workflow.enabled || isTriggering || isDeleting}
            title={
              !workflow.enabled
                ? t('dashboard.card.runNowDisabledPaused')
                : undefined
            }
            onClick={() => onTrigger(workflow)}
          >
            <Play aria-hidden="true" />
            {isTriggering
              ? t('dashboard.card.running')
              : t('dashboard.card.runNow')}
          </Button>

          <Button variant="ghost" size="sm" asChild>
            <Link to={`/workflows/${workflow.id}/edit`}>
              <Pencil aria-hidden="true" />
              {t('dashboard.card.edit')}
            </Link>
          </Button>
        </div>

        {!confirmingDelete && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={isDeleting}
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => setConfirmingDelete(true)}
          >
            <Trash2 aria-hidden="true" />
            {t('dashboard.card.delete')}
          </Button>
        )}
      </div>

      {confirmingDelete && (
        <div
          role="region"
          aria-label={t('dashboard.card.deleteConfirmTitle', {
            name: workflow.name,
          })}
          className="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-md border border-destructive/30 bg-destructive/5 p-3"
        >
          <div className="space-y-0.5 text-sm">
            <p className="font-medium text-destructive">
              {t('dashboard.card.deleteConfirmTitle', { name: workflow.name })}
            </p>
            <p className="text-xs text-muted-foreground">
              {t('dashboard.card.deleteConfirmDescription')}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isDeleting}
              onClick={() => setConfirmingDelete(false)}
            >
              {t('dashboard.card.deleteCancel')}
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              disabled={isDeleting}
              onClick={() => onDelete(workflow)}
            >
              {isDeleting
                ? t('dashboard.card.deleting')
                : t('dashboard.card.deleteConfirm')}
            </Button>
          </div>
        </div>
      )}
    </article>
  );
}
