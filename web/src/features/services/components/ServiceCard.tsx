import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import type { AboutService, LinkedService } from '@/shared/api/generated/model';
import {
  getListServicesQueryKey,
  useAuthorizeService,
  useUnlinkService,
} from '@/shared/api/generated/services/services';
import { Button } from '@/shared/ui/button';
import { serviceErrorKey } from '../lib/errors';
import { goTo } from '../lib/redirect';

export function ServiceCard({
  service,
  link,
}: {
  service: AboutService;
  link?: LinkedService;
}) {
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: getListServicesQueryKey() });

  const authorize = useAuthorizeService({
    mutation: { onSuccess: ({ url }) => goTo(url) },
  });
  const unlink = useUnlinkService({ mutation: { onSettled: refresh } });

  const error = authorize.error
    ? serviceErrorKey(authorize.error, 'link')
    : unlink.error && serviceErrorKey(unlink.error, 'unlink');

  return (
    <li className="flex flex-col gap-3 rounded-lg border p-4">
      <div>
        <h2 className="text-lg font-medium">{service.name}</h2>
        <p className="text-muted-foreground text-sm">{service.description}</p>
      </div>

      {link ? (
        <div className="text-sm">
          <p>
            {t('services.linkedOn', {
              date: new Date(link.linked_at).toLocaleDateString(i18n.language, {
                dateStyle: 'medium',
              }),
            })}
          </p>
          {link.scopes && (
            <p className="text-muted-foreground">
              {t('services.scopes', { scopes: link.scopes })}
            </p>
          )}
        </div>
      ) : (
        <p className="text-sm">{t('services.notLinked')}</p>
      )}

      {error && (
        <p role="alert" className="text-destructive text-sm">
          {t(error)}
        </p>
      )}

      <div className="mt-auto">
        {link ? (
          <Button
            variant="outline"
            onClick={() => unlink.mutate({ service: service.name })}
            disabled={unlink.isPending}
          >
            {unlink.isPending
              ? t('services.unlinking')
              : t('services.unlink', { service: service.name })}
          </Button>
        ) : (
          <Button
            onClick={() => authorize.mutate({ service: service.name })}
            disabled={authorize.isPending}
          >
            {authorize.isPending
              ? t('services.linking', { service: service.name })
              : t('services.link', { service: service.name })}
          </Button>
        )}
      </div>
    </li>
  );
}
