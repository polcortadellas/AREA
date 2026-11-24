import { z } from 'zod';

// Messages are i18n keys; FormMessage translates them. Limits match the server's.
const email = z.email('validation.email');

export const loginSchema = z.object({
  email,
  password: z.string().min(1, 'validation.required'),
});
export type LoginValues = z.infer<typeof loginSchema>;

export const registerSchema = z
  .object({
    email,
    password: z
      .string()
      .min(8, 'validation.passwordLength')
      .max(128, 'validation.passwordLength'),
    confirmPassword: z.string().min(1, 'validation.required'),
  })
  .refine(v => v.password === v.confirmPassword, {
    message: 'validation.passwordMismatch',
    path: ['confirmPassword'],
  });
export type RegisterValues = z.infer<typeof registerSchema>;
