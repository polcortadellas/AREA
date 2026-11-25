import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';

import type { Run } from '@/shared/api/generated/model';
import { cn } from '@/shared/lib/utils';
import { formatDateTime, formatDuration, shortId } from '../lib/format';
import { RunStatusBadge } from './StatusBadge';

interface RunListItemProps {
  workflowId: string;
  run: Run;
  isSelected: boolean;
}

export function RunListItem({ workflowId, run, isSelected }: RunListItemProps) {
  const { t, i18n } = useTranslation();
  const created = formatDateTime(run.created_at, i18n.language);
  const duration = formatDuration(run.started_at, run.finished_at);

  return (
    <li>
      <Link
        to={`/workflows/${workflowId}/runs/${run.id}`}
        aria-current={isSelected ? 'page' : undefined}
        className={cn(
          'block rounded-lg border p-3.5 transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50',
          isSelected
            ? 'border-primary bg-accent/60 shadow-xs'
            : 'bg-card hover:bg-accent/40',
        )}
      >
        <div className="flex items-center justify-between gap-2">
          <RunStatusBadge status={run.status} />
          <code className="text-xs text-muted-foreground">
            #{shortId(run.id)}
          </code>
        </div>

        <div className="mt-2 flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
          <span>{created ?? t('dashboard.history.notAvailable')}</span>
          {duration && <span>{duration}</span>}
        </div>

        {run.error && (
          <p className="mt-1.5 truncate text-xs text-destructive">
            {run.error}
          </p>
        )}
      </Link>
    </li>
  );
}
