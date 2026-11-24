import type { HTMLInputAutoCompleteAttribute } from 'react';
import type { Control, FieldPath, FieldValues } from 'react-hook-form';

import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/shared/ui/form';
import { Input } from '@/shared/ui/input';

export function InputField<T extends FieldValues>({
  control,
  name,
  label,
  type,
  autoComplete,
  description,
}: {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  type: 'email' | 'password';
  autoComplete: HTMLInputAutoCompleteAttribute;
  description?: string;
}) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input type={type} autoComplete={autoComplete} {...field} />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
