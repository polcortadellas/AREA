import { useTranslation } from 'react-i18next';

export function StepErrorAlert({
  nodeId,
  stepLabel,
  problems,
}: {
  nodeId: string;
  stepLabel: string;
  problems?: ReadonlyArray<string>;
}) {
  const { t } = useTranslation();

  if (!problems || problems.length === 0) {
    return null;
  }

  return (
    <div
      id={`step-error-${nodeId}`}
      data-testid={`step-error-${nodeId}`}
      role="alert"
      className="border-destructive/50 bg-destructive/10 text-destructive rounded-md border p-3 text-sm"
    >
      <p className="font-medium">
        {t('builder.errors.stepProblems', { step: stepLabel })}
      </p>
      <ul className="mt-1 list-inside list-disc space-y-0.5">
        {problems.map((problem, index) => (
          <li key={`${nodeId}-problem-${index}`}>{problem}</li>
        ))}
      </ul>
    </div>
  );
}
