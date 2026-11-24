import { useTranslation } from 'react-i18next';

import {
  appendKeyToken,
  type UpstreamOutputOption,
} from '@/features/builder/lib/catalogue';
import {
  CONDITION_OPERATORS,
  type ConditionOperator,
  type DraftCondition,
} from '@/features/builder/lib/workflowGraph';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { StepErrorAlert } from './StepErrorAlert';

export function ConditionStepCard({
  condition,
  upstreamOutputs,
  problems,
  fieldErrors,
  onChange,
  onRemove,
}: {
  condition: DraftCondition;
  upstreamOutputs: ReadonlyArray<UpstreamOutputOption>;
  problems?: ReadonlyArray<string>;
  fieldErrors?: Record<string, string[]>;
  onChange: (next: DraftCondition) => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation();
  const hasError = Boolean(problems && problems.length > 0);
  const leftHasError = Boolean(fieldErrors?.left?.length);
  const rightHasError = Boolean(fieldErrors?.right?.length);
  const stepErrorId = `step-error-${condition.id}`;

  return (
    <section
      aria-labelledby="condition-step-heading"
      data-testid={`step-${condition.id}`}
      data-node-id={condition.id}
      className={cn(
        'space-y-4 rounded-lg border p-5 shadow-xs',
        hasError && 'border-destructive',
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h2 id="condition-step-heading" className="text-lg font-semibold">
            {t('builder.condition.heading')}
          </h2>
          <p className="text-muted-foreground text-sm">
            {t('builder.condition.description')}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="bg-muted rounded px-2 py-0.5 font-mono text-xs">
            {t('builder.nodeIdLabel', { id: condition.id })}
          </span>
          <Button type="button" variant="outline" size="sm" onClick={onRemove}>
            {t('builder.condition.remove')}
          </Button>
        </div>
      </div>

      <StepErrorAlert
        nodeId={condition.id}
        stepLabel={t('builder.condition.heading')}
        problems={problems}
      />

      <div className="grid gap-4 md:grid-cols-3">
        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Label htmlFor="condition-left">
              {t('builder.condition.leftLabel')}
            </Label>
            {upstreamOutputs.length > 0 && (
              <select
                aria-label={t('builder.keyMap.insertLabel', { field: 'left' })}
                value=""
                onChange={e => {
                  if (!e.target.value) return;
                  onChange({
                    ...condition,
                    left: appendKeyToken(condition.left, e.target.value),
                  });
                }}
                className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-8 rounded-md border px-2 text-xs outline-none focus-visible:ring-[3px]"
              >
                <option value="">{t('builder.keyMap.placeholder')}</option>
                {upstreamOutputs.map(opt => (
                  <option key={opt.token} value={opt.token}>
                    {opt.token} ({opt.field.type})
                  </option>
                ))}
              </select>
            )}
          </div>
          <Input
            id="condition-left"
            type="text"
            value={condition.left}
            onChange={e => onChange({ ...condition, left: e.target.value })}
            aria-describedby={
              leftHasError
                ? `condition-left-desc ${stepErrorId}`
                : 'condition-left-desc'
            }
            aria-invalid={leftHasError}
          />
          <p id="condition-left-desc" className="text-muted-foreground text-xs">
            {t('builder.condition.leftDescription')}
          </p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="condition-op">{t('builder.condition.opLabel')}</Label>
          <select
            id="condition-op"
            value={condition.op}
            onChange={e =>
              onChange({
                ...condition,
                op: e.target.value as ConditionOperator,
              })
            }
            className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border px-3 text-sm outline-none focus-visible:ring-[3px]"
          >
            {CONDITION_OPERATORS.map(op => (
              <option key={op} value={op}>
                {t(`builder.condition.operators.${op}`)}
              </option>
            ))}
          </select>
        </div>

        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Label htmlFor="condition-right">
              {t('builder.condition.rightLabel')}
            </Label>
            {upstreamOutputs.length > 0 && (
              <select
                aria-label={t('builder.keyMap.insertLabel', { field: 'right' })}
                value=""
                onChange={e => {
                  if (!e.target.value) return;
                  onChange({
                    ...condition,
                    right: appendKeyToken(condition.right, e.target.value),
                  });
                }}
                className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-8 rounded-md border px-2 text-xs outline-none focus-visible:ring-[3px]"
              >
                <option value="">{t('builder.keyMap.placeholder')}</option>
                {upstreamOutputs.map(opt => (
                  <option key={opt.token} value={opt.token}>
                    {opt.token} ({opt.field.type})
                  </option>
                ))}
              </select>
            )}
          </div>
          <Input
            id="condition-right"
            type="text"
            value={condition.right}
            onChange={e => onChange({ ...condition, right: e.target.value })}
            aria-describedby={
              rightHasError
                ? `condition-right-desc ${stepErrorId}`
                : 'condition-right-desc'
            }
            aria-invalid={rightHasError}
          />
          <p
            id="condition-right-desc"
            className="text-muted-foreground text-xs"
          >
            {t('builder.condition.rightDescription')}
          </p>
        </div>
      </div>
    </section>
  );
}
