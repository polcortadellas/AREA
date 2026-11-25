import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router';

import { useListServices } from '@/shared/api/generated/services/services';
import { useGetAbout } from '@/shared/api/generated/system/system';
import { ServiceCard } from '../components/ServiceCard';
import { callbackErrorKey } from '../lib/errors';

/**
 * The OAuth callback redirects here with ?linked=<service> or ?error=<code>.
 * The result is kept in state and removed from the URL, so a reload does not
 * show it again.
 */
function useCallbackResult() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [result] = useState(() => ({
    linked: searchParams.get('linked'),
    error: searchParams.get('error'),
  }));

  useEffect(() => {
    if (searchParams.has('linked') || searchParams.has('error')) {
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  return result;
}

export function ServicesPage() {
  const { t } = useTranslation();
  const about = useGetAbout();
  const linked = useListServices();
  const result = useCallbackResult();

  // Only services that act on a user's own account need linking.
  const services = about.data?.server.services.filter(s => s.oauth2) ?? [];

  let content;
  if (about.isPending || linked.isPending) {
    content = (
      <p role="status" className="text-muted-foreground">
        {t('layout.loading')}
      </p>
    );
  } else if (about.isError || linked.isError) {
    content = (
      <p role="alert" className="text-destructive">
        {t('services.loadError')}
      </p>
    );
  } else if (services.length === 0) {
    content = <p className="text-muted-foreground">{t('services.empty')}</p>;
  } else {
    content = (
      <ul className="grid gap-4 sm:grid-cols-2">
        {services.map(service => (
          <ServiceCard
            key={service.name}
            service={service}
            link={linked.data.find(l => l.service === service.name)}
          />
        ))}
      </ul>
    );
  }

  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">{t('services.title')}</h1>
        <p className="text-muted-foreground">{t('services.description')}</p>
      </div>

      {result.linked && (
        <p role="status" className="rounded-md border p-3 text-sm">
          {t('services.callback.linked', { service: result.linked })}
        </p>
      )}
      {result.error && (
        <p
          role="alert"
          className="border-destructive text-destructive rounded-md border p-3 text-sm"
        >
          {t(callbackErrorKey(result.error))}
        </p>
      )}

      {content}
    </section>
  );
}
