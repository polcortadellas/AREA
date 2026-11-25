import {
  AlertOctagon,
  CheckCircle2,
  Clock,
  Loader2,
  PauseCircle,
  SkipForward,
  XCircle,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { RunStatus, RunStepStatus } from '@/shared/api/generated/model';
import { cn } from '@/shared/lib/utils';

const badgeBase =
  'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium';

export function WorkflowStatusBadge({ enabled }: { enabled: boolean }) {
  const { t } = useTranslation();

  return (
    <span
      className={cn(
        badgeBase,
        enabled
          ? 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-800/60 dark:bg-emerald-950/60 dark:text-emerald-300'
          : 'border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-800/60 dark:bg-amber-950/60 dark:text-amber-300',
      )}
    >
      {enabled ? (
        <CheckCircle2 className="size-3.5" aria-hidden="true" />
      ) : (
        <PauseCircle className="size-3.5" aria-hidden="true" />
      )}
      {enabled ? t('dashboard.status.active') : t('dashboard.status.paused')}
    </span>
  );
}

const runStatusClasses: Record<RunStatus, string> = {
  PENDING:
    'border-slate-200 bg-slate-50 text-slate-800 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200',
  RUNNING:
    'border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-800/60 dark:bg-blue-950/60 dark:text-blue-300',
  SUCCEEDED:
    'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-800/60 dark:bg-emerald-950/60 dark:text-emerald-300',
  FAILED:
    'border-red-200 bg-red-50 text-red-800 dark:border-red-800/60 dark:bg-red-950/60 dark:text-red-300',
  DEAD: 'border-red-300 bg-red-100 text-red-900 dark:border-red-800 dark:bg-red-950 dark:text-red-200',
};

export function RunStatusBadge({ status }: { status: RunStatus }) {
  const { t } = useTranslation();

  return (
    <span className={cn(badgeBase, runStatusClasses[status])}>
      {status === 'PENDING' && (
        <Clock className="size-3.5" aria-hidden="true" />
      )}
      {status === 'RUNNING' && (
        <Loader2 className="size-3.5 animate-spin" aria-hidden="true" />
      )}
      {status === 'SUCCEEDED' && (
        <CheckCircle2 className="size-3.5" aria-hidden="true" />
      )}
      {status === 'FAILED' && (
        <XCircle className="size-3.5" aria-hidden="true" />
      )}
      {status === 'DEAD' && (
        <AlertOctagon className="size-3.5" aria-hidden="true" />
      )}
      {t(`dashboard.history.runStatus.${status}`)}
    </span>
  );
}

const stepStatusClasses: Record<RunStepStatus, string> = {
  PENDING:
    'border-slate-200 bg-slate-50 text-slate-800 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200',
  RUNNING:
    'border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-800/60 dark:bg-blue-950/60 dark:text-blue-300',
  SUCCEEDED:
    'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-800/60 dark:bg-emerald-950/60 dark:text-emerald-300',
  FAILED:
    'border-red-200 bg-red-50 text-red-800 dark:border-red-800/60 dark:bg-red-950/60 dark:text-red-300',
  SKIPPED:
    'border-zinc-200 bg-zinc-100 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-300',
};

export function StepStatusBadge({ status }: { status: RunStepStatus }) {
  const { t } = useTranslation();

  return (
    <span className={cn(badgeBase, stepStatusClasses[status])}>
      {status === 'PENDING' && (
        <Clock className="size-3.5" aria-hidden="true" />
      )}
      {status === 'RUNNING' && (
        <Loader2 className="size-3.5 animate-spin" aria-hidden="true" />
      )}
      {status === 'SUCCEEDED' && (
        <CheckCircle2 className="size-3.5" aria-hidden="true" />
      )}
      {status === 'FAILED' && (
        <XCircle className="size-3.5" aria-hidden="true" />
      )}
      {status === 'SKIPPED' && (
        <SkipForward className="size-3.5" aria-hidden="true" />
      )}
      {t(`dashboard.history.stepStatus.${status}`)}
    </span>
  );
}
