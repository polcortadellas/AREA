import { zodResolver } from '@hookform/resolvers/zod';
import { screen } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import { describe, expect, it, vi } from 'vitest';
import { z } from 'zod';

import { i18n } from '@/shared/i18n';
import { expectNoA11yViolations } from '@/test/axe';
import { renderWithProviders } from '@/test/render';
import { Button } from './button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from './form';
import { Input } from './input';

// Messages are i18n keys; FormMessage translates them.
const schema = z.object({
  email: z.email('validation.email'),
});

function ExampleForm({
  onSubmit,
}: {
  onSubmit: (v: { email: string }) => void;
}) {
  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: { email: '' },
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} noValidate>
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input type="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Send</Button>
      </form>
    </Form>
  );
}

describe('Form', () => {
  it('shows a translated, accessible error when validation fails', async () => {
    const onSubmit = vi.fn();
    const { user, container } = renderWithProviders(
      <ExampleForm onSubmit={onSubmit} />,
    );

    await user.click(screen.getByRole('button', { name: 'Send' }));

    const input = screen.getByLabelText('Email');
    const error = await screen.findByRole('alert');
    expect(error).toHaveTextContent('Enter a valid email address');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(input.getAttribute('aria-describedby')).toContain(error.id);
    expect(onSubmit).not.toHaveBeenCalled();
    await expectNoA11yViolations(container);

    await i18n.changeLanguage('es');
    expect(
      await screen.findByText('Introduce un correo electrónico válido'),
    ).toBeInTheDocument();
  });

  it('submits valid values', async () => {
    const onSubmit = vi.fn();
    const { user } = renderWithProviders(<ExampleForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Email'), 'me@example.com');
    await user.click(screen.getByRole('button', { name: 'Send' }));

    expect(onSubmit).toHaveBeenCalledWith(
      { email: 'me@example.com' },
      expect.anything(),
    );
  });
});
