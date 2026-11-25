import { useTranslation } from 'react-i18next';

import {
  appendKeyToken,
  type UpstreamOutputOption,
} from '@/features/builder/lib/catalogue';
import type { DraftConfig } from '@/features/builder/lib/workflowGraph';
import type { Field } from '@/shared/api/generated/model';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';

export function NodeFieldsForm({
  nodeId,
  inputs,
  values,
  fieldErrors,
  upstreamOutputs = [],
  onChange,
}: {
  nodeId: string;
  inputs: ReadonlyArray<Field>;
  values: DraftConfig;
  fieldErrors?: Record<string, string[]>;
  upstreamOutputs?: ReadonlyArray<UpstreamOutputOption>;
  onChange: (key: string, value: string | boolean) => void;
}) {
  const { t } = useTranslation();

  if (inputs.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">
        {t('builder.reactions.noInputs')}
      </p>
    );
  }

  return (
    <div className="space-y-4">
      {inputs.map(field => {
        const controlId = `${nodeId}-field-${field.key}`;
        const descId = `${controlId}-desc`;
        const errors = fieldErrors?.[field.key] ?? [];
        const hasError = errors.length > 0;
        const stepErrorId = `step-error-${nodeId}`;
        const describedBy = hasError ? `${descId} ${stepErrorId}` : descId;

        if (field.type === 'boolean') {
          return (
            <div key={field.key} className="space-y-1.5">
              <div className="flex items-center gap-2">
                <input
                  id={controlId}
                  type="checkbox"
                  checked={Boolean(values[field.key])}
                  onChange={e => onChange(field.key, e.target.checked)}
                  aria-describedby={describedBy}
                  aria-invalid={hasError}
                  className="size-4 rounded border"
                />
                <Label htmlFor={controlId}>{field.key}</Label>
                <span className="text-muted-foreground text-xs">
                  ({field.type} ·{' '}
                  {field.required
                    ? t('builder.requiredField')
                    : t('builder.optionalField')}
                  )
                </span>
              </div>
              <p id={descId} className="text-muted-foreground text-xs">
                {field.description}
              </p>
            </div>
          );
        }

        const stringVal =
          typeof values[field.key] === 'string'
            ? (values[field.key] as string)
            : '';

        return (
          <div key={field.key} className="space-y-1.5">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-center gap-1.5">
                <Label htmlFor={controlId}>{field.key}</Label>
                <span className="text-muted-foreground text-xs">
                  ({field.type} ·{' '}
                  {field.required
                    ? t('builder.requiredField')
                    : t('builder.optionalField')}
                  )
                </span>
              </div>

              {upstreamOutputs.length > 0 && (
                <select
                  aria-label={t('builder.keyMap.insertLabel', {
                    field: field.key,
                  })}
                  value=""
                  onChange={e => {
                    if (!e.target.value) return;
                    onChange(
                      field.key,
                      appendKeyToken(stringVal, e.target.value),
                    );
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
              id={controlId}
              type="text"
              inputMode={field.type === 'number' ? 'decimal' : 'text'}
              value={stringVal}
              onChange={e => onChange(field.key, e.target.value)}
              aria-describedby={describedBy}
              aria-invalid={hasError}
            />

            <p id={descId} className="text-muted-foreground text-xs">
              {field.description}
            </p>
          </div>
        );
      })}
    </div>
  );
}

export function DeclaredOutputsList({
  referencePrefix,
  outputs,
}: {
  referencePrefix: string;
  outputs: ReadonlyArray<Field>;
}) {
  const { t } = useTranslation();

  return (
    <div className="bg-muted/40 rounded-md border px-3 py-2 text-xs">
      <p className="font-medium">{t('builder.outputs.declaredHeading')}</p>
      {outputs.length === 0 ? (
        <p className="text-muted-foreground mt-1">
          {t('builder.outputs.none')}
        </p>
      ) : (
        <ul className="mt-1 flex flex-wrap gap-2">
          {outputs.map(output => (
            <li
              key={output.key}
              title={output.description}
              className="bg-background inline-flex items-center gap-1 rounded border px-2 py-0.5 font-mono"
            >
              <span>{`{{ ${referencePrefix}.${output.key} }}`}</span>
              <span className="text-muted-foreground font-sans">
                ({output.type})
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
