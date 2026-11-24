import { useTranslation } from 'react-i18next';

import {
  findNodeByType,
  findServiceByName,
  reactionServices,
  type UpstreamOutputOption,
} from '@/features/builder/lib/catalogue';
import type { DraftReaction } from '@/features/builder/lib/workflowGraph';
import type { About } from '@/shared/api/generated/model';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/shared/ui/button';
import { Label } from '@/shared/ui/label';
import { DeclaredOutputsList, NodeFieldsForm } from './NodeFieldsForm';
import { StepErrorAlert } from './StepErrorAlert';

export function ReactionStepCard({
  about,
  reaction,
  stepTitle,
  upstreamOutputs,
  problems,
  fieldErrors,
  canMoveUp,
  canMoveDown,
  canRemove,
  onServiceChange,
  onReactionChange,
  onConfigChange,
  onMoveUp,
  onMoveDown,
  onRemove,
}: {
  about: About;
  reaction: DraftReaction;
  stepTitle: string;
  upstreamOutputs: ReadonlyArray<UpstreamOutputOption>;
  problems?: ReadonlyArray<string>;
  fieldErrors?: Record<string, string[]>;
  canMoveUp: boolean;
  canMoveDown: boolean;
  canRemove: boolean;
  onServiceChange: (serviceName: string) => void;
  onReactionChange: (reactionType: string) => void;
  onConfigChange: (key: string, value: string | boolean) => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation();
  const availableServices = reactionServices(about);
  const selectedService =
    findServiceByName(about, reaction.service) ?? availableServices[0];
  const selectedReaction =
    findNodeByType(about, 'reaction', reaction.type) ??
    selectedService?.reactions[0];
  const hasError = Boolean(problems && problems.length > 0);
  const headingId = `heading-${reaction.id}`;
  const serviceSelectId = `${reaction.id}-service`;
  const reactionSelectId = `${reaction.id}-reaction`;

  return (
    <section
      aria-labelledby={headingId}
      data-testid={`step-${reaction.id}`}
      data-node-id={reaction.id}
      className={cn(
        'space-y-4 rounded-lg border p-5 shadow-xs',
        hasError && 'border-destructive',
      )}
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <h3 id={headingId} className="text-base font-semibold">
            {stepTitle}
          </h3>
          <span className="bg-muted rounded px-2 py-0.5 font-mono text-xs">
            {t('builder.nodeIdLabel', { id: reaction.id })}
          </span>
        </div>

        <div className="flex items-center gap-1.5">
          {canMoveUp && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onMoveUp}
              aria-label={`${t('builder.reactions.moveUp')} ${stepTitle}`}
            >
              {t('builder.reactions.moveUp')}
            </Button>
          )}
          {canMoveDown && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onMoveDown}
              aria-label={`${t('builder.reactions.moveDown')} ${stepTitle}`}
            >
              {t('builder.reactions.moveDown')}
            </Button>
          )}
          {canRemove && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onRemove}
              aria-label={`${t('builder.reactions.remove')} ${stepTitle}`}
            >
              {t('builder.reactions.remove')}
            </Button>
          )}
        </div>
      </div>

      <StepErrorAlert
        nodeId={reaction.id}
        stepLabel={stepTitle}
        problems={problems}
      />

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between gap-2">
            <Label htmlFor={serviceSelectId}>{t('builder.serviceLabel')}</Label>
            {selectedService?.oauth2 && (
              <span className="bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs font-medium">
                {t('builder.oauth2Badge')}
              </span>
            )}
          </div>
          <select
            id={serviceSelectId}
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
          <Label htmlFor={reactionSelectId}>
            {t('builder.reactions.reactionLabel')}
          </Label>
          <select
            id={reactionSelectId}
            value={selectedReaction?.type ?? ''}
            onChange={e => onReactionChange(e.target.value)}
            className="border-input bg-background focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border px-3 text-sm outline-none focus-visible:ring-[3px]"
          >
            {(selectedService?.reactions ?? []).map(node => (
              <option key={node.type} value={node.type}>
                {node.name}
              </option>
            ))}
          </select>
          {selectedReaction?.description && (
            <p className="text-muted-foreground text-xs">
              {selectedReaction.description}
            </p>
          )}
        </div>
      </div>

      <NodeFieldsForm
        nodeId={reaction.id}
        inputs={selectedReaction?.inputs ?? []}
        values={reaction.config}
        fieldErrors={fieldErrors}
        upstreamOutputs={upstreamOutputs}
        onChange={onConfigChange}
      />

      <DeclaredOutputsList
        referencePrefix={reaction.id}
        outputs={selectedReaction?.outputs ?? []}
      />
    </section>
  );
}
