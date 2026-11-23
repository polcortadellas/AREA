import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';

import { useRegister } from '@/shared/api/generated/auth/auth';
import { Button } from '@/shared/ui/button';
import { Form } from '@/shared/ui/form';
import { AuthCard, FormAlert } from '../components/AuthCard';
import { InputField } from '../components/InputField';
import { authErrorKey } from '../lib/errors';
import { registerSchema, type RegisterValues } from '../lib/schemas';

export function RegisterPage() {
  const { t } = useTranslation();
  const register = useRegister();
  const form = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: '', password: '', confirmPassword: '' },
  });

  const onSubmit = form.handleSubmit(({ email, password }) =>
    register.mutate({ data: { email, password } }),
  );

  if (register.isSuccess) {
    return (
      <AuthCard title={t('auth.checkEmail.title')}>
        <p>
          {t('auth.checkEmail.description', {
            email: register.variables.data.email,
          })}
        </p>
        <p className="text-muted-foreground text-sm">
          {t('auth.checkEmail.existing')}
        </p>
        <Button asChild variant="outline">
          <Link to="/login">{t('auth.checkEmail.login')}</Link>
        </Button>
      </AuthCard>
    );
  }

  return (
    <AuthCard title={t('auth.register.title')}>
      <Form {...form}>
        <form onSubmit={onSubmit} noValidate className="space-y-4">
          <InputField
            control={form.control}
            name="email"
            label={t('auth.email')}
            type="email"
            autoComplete="email"
          />
          <InputField
            control={form.control}
            name="password"
            label={t('auth.password')}
            type="password"
            autoComplete="new-password"
            description={t('auth.passwordHint')}
          />
          <InputField
            control={form.control}
            name="confirmPassword"
            label={t('auth.confirmPassword')}
            type="password"
            autoComplete="new-password"
          />
          <FormAlert>
            {register.error && t(authErrorKey(register.error))}
          </FormAlert>
          <Button
            type="submit"
            className="w-full"
            disabled={register.isPending}
          >
            {register.isPending
              ? t('auth.register.submitting')
              : t('auth.register.submit')}
          </Button>
        </form>
      </Form>
      <p className="text-muted-foreground text-sm">
        {t('auth.register.haveAccount')}{' '}
        <Link to="/login" className="text-foreground underline">
          {t('auth.register.login')}
        </Link>
      </p>
    </AuthCard>
  );
}
