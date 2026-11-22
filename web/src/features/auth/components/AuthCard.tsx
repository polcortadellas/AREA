import type { ReactNode } from 'react';

export function AuthCard({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="mx-auto w-full max-w-sm space-y-6 py-8">
      <h1 className="text-2xl font-semibold">{title}</h1>
      {children}
    </section>
  );
}

/** Request error shown above a form's submit button, announced by screen readers. */
export function FormAlert({ children }: { children?: ReactNode }) {
  if (!children) return null;
  return (
    <p role="alert" className="text-destructive text-sm">
      {children}
    </p>
  );
}
