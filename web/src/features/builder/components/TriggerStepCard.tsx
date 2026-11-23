import { useTranslation } from 'react-i18next';

import {
  findNodeByType,
  findServiceByName,
  triggerServices,
} from '@/features/builder/lib/catalogue';
import type { DraftTrigger } from '@/features/builder/lib/workflowGraph';
import type { About } from '@/shared/api/generated/model';
import { cn } from '@/shared/lib/utils';
import { Label } from '@/shared/ui/label';
import { DeclaredOutputsList, NodeFieldsForm } from './NodeFieldsForm';
import { StepErrorAlert } from './StepErrorAlert';

export function TriggerStepCard({
  about,
  trigger,
  problems,
  fieldErrors,
  onServiceChange,
  onActionChange,
  onConfigChange,
}: {
  about: About;
  trigger: DraftTrigger;
  problems?: ReadonlyArray<string>;
  fieldErrors?: Record<string, string[]>;
  onServiceChange: (serviceName: string) => void;
  onActionChange: (actionType: string) => void;
  onConfigChange: (key: string, value: string | boolean) => void;
}) {
  const { t } = useTranslation();
  const availableServices = triggerServices(about);
  const selectedService =
    findServiceByName(about, trigger.service) ?? availableServices[0];
  const selectedAction =
    findNodeByType(about, 'trigger', trigger.type) ??
    selectedService?.actions[0];
  const hasError = Boolean(problems && problems.length > 0);

  return (
    <section
      aria-labelledby="trigger-step-heading"
      data-testid={`step-${trigger.id}`}
      data-node-id={trigger.id}
      className={cn(
        'space-y-4 rounded-lg border p-5 shadow-xs',
        hasError && 'border-destructive',
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h2 id="trigger-step-heading" className="text-lg font-semibold">
            {t('builder.trigger.heading')}
          </h2>
          <p className="text-muted-foreground text-sm">
            {t('builder.trigger.description')}
          </p>
        </div>
        <span className="bg-muted rounded px-2 py-0.5 font-mono text-xs">
          {t('builder.nodeIdLabel', { id: trigger.id })}
        </span>
      </div>

      <StepErrorAlert
        nodeId={trigger.id}
        stepLabel={t('builder.trigger.heading')}
        problems={problems}
      />

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between gap-2">
            <Label htmlFor="trigger-service">{t('builder.serviceLabel')}</Label>
            {selectedService?.oauth2 && (
              <span className="bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs font-medium">
                {t('builder.oauth2Badge')}
              </span>
            )}
          </div>
          <select
            id="trigger-service"
            value={selectedService?.name ?? ''}
            onChange={e => onServiceChange(e.target.value)}
            className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border px-3 text-sm outline-none focus-visible:ring-[3px]"
          >
            {availableServices.map(service => (
              <option key={service.name} value={service.name}>
                {service.name}
              </option>
            ))}
          </select>
          {selectedService?.description && (
            <p className="text-muted-foreground text-xs">
              {selectedService.description}
            </p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="trigger-action">
            {t('builder.trigger.actionLabel')}
          </Label>
          <select
            id="trigger-action"
            value={selectedAction?.type ?? ''}
            onChange={e => onActionChange(e.target.value)}
            className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border px-3 text-sm outline-none focus-visible:ring-[3px]"
          >
            {(selectedService?.actions ?? []).map(action => (
              <option key={action.type} value={action.type}>
                {action.name}
              </option>
            ))}
          </select>
          {selectedAction?.description && (
            <p className="text-muted-foreground text-xs">
              {selectedAction.description}
            </p>
          )}
        </div>
      </div>

      <NodeFieldsForm
        nodeId={trigger.id}
        inputs={selectedAction?.inputs ?? []}
        values={trigger.config}
        fieldErrors={fieldErrors}
        onChange={onConfigChange}
      />

      <DeclaredOutputsList
        referencePrefix="trigger"
        outputs={selectedAction?.outputs ?? []}
      />
    </section>
  );
}
