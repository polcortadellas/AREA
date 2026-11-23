import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useSearchParams } from 'react-router';

import { useVerifyEmail } from '@/shared/api/generated/auth/auth';
import { Button } from '@/shared/ui/button';
import { AuthCard, FormAlert } from '../components/AuthCard';
import { useStartSession } from '../hooks/useSessionActions';
import { authErrorKey } from '../lib/errors';

/**
 * Target of the link in the confirmation e-mail (/verify-email?token=...).
 * The token is only sent when the user clicks the button, so mail scanners
 * that open links cannot use it up.
 */
export function VerifyEmailPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const startSession = useStartSession();
  const verify = useVerifyEmail();
  const token = searchParams.get('token');

  const confirm = () => {
    if (!token) return;
    verify.mutate(
      { data: { token } },
      {
        onSuccess: async () => {
          await startSession();
          await navigate('/', { replace: true });
        },
      },
    );
  };

  const registerAgain = (
    <Button asChild variant="outline">
      <Link to="/register">{t('auth.verify.register')}</Link>
    </Button>
  );

  if (!token) {
    return (
      <AuthCard title={t('auth.verify.title')}>
        <FormAlert>{t('auth.verify.missingToken')}</FormAlert>
        {registerAgain}
      </AuthCard>
    );
  }

  return (
    <AuthCard title={t('auth.verify.title')}>
      <p>{t('auth.verify.description')}</p>
      <FormAlert>{verify.error && t(authErrorKey(verify.error))}</FormAlert>
      {verify.error ? (
        registerAgain
      ) : (
        <Button
          className="w-full"
          onClick={confirm}
          disabled={verify.isPending}
        >
          {verify.isPending
            ? t('auth.verify.submitting')
            : t('auth.verify.submit')}
        </Button>
      )}
    </AuthCard>
  );
}
