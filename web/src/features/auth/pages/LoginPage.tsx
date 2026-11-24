import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useSearchParams } from 'react-router';

import { useLogin } from '@/shared/api/generated/auth/auth';
import { Button } from '@/shared/ui/button';
import { Form } from '@/shared/ui/form';
import { AuthCard, FormAlert } from '../components/AuthCard';
import { InputField } from '../components/InputField';
import { useStartSession } from '../hooks/useSessionActions';
import { authErrorKey } from '../lib/errors';
import { safeNext } from '../lib/next';
import { loginSchema, type LoginValues } from '../lib/schemas';

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const startSession = useStartSession();
  const login = useLogin();
  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const onSubmit = form.handleSubmit(values =>
    login.mutate(
      { data: values },
      {
        onSuccess: async () => {
          await startSession();
          await navigate(safeNext(searchParams.get('next')), { replace: true });
        },
      },
    ),
  );

  return (
    <AuthCard title={t('auth.login.title')}>
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
            autoComplete="current-password"
          />
          <FormAlert>{login.error && t(authErrorKey(login.error))}</FormAlert>
          <Button type="submit" className="w-full" disabled={login.isPending}>
            {login.isPending
              ? t('auth.login.submitting')
              : t('auth.login.submit')}
          </Button>
        </form>
      </Form>
      <p className="text-muted-foreground text-sm">
        {t('auth.login.noAccount')}{' '}
        <Link to="/register" className="text-foreground underline">
          {t('auth.login.register')}
        </Link>
      </p>
    </AuthCard>
  );
}
