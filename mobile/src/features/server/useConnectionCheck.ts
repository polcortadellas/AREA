import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { checkConnection, type ConnectionResult } from './checkConnection';

export type ConnectionStatus =
  | { state: 'idle' }
  | { state: 'checking' }
  | { state: 'done'; result: ConnectionResult };

/** Runs checkConnection and keeps only the answer for the latest request. */
export function useConnectionCheck() {
  const [status, setStatus] = useState<ConnectionStatus>({ state: 'idle' });
  const latest = useRef(0);

  useEffect(
    () => () => {
      latest.current = -1;
    },
    [],
  );

  const check = useCallback(async (baseUrl: string) => {
    const id = ++latest.current;
    setStatus({ state: 'checking' });
    const result = await checkConnection(baseUrl);
    if (id === latest.current) {
      setStatus({ state: 'done', result });
    }
    return result;
  }, []);

  const reset = useCallback(() => {
    latest.current++;
    setStatus({ state: 'idle' });
  }, []);

  return { status, check, reset };
}

/** Translated message and tone for a connection status, or null when idle. */
export function useConnectionMessage(status: ConnectionStatus) {
  const { t } = useTranslation();
  if (status.state === 'idle') {
    return null;
  }
  if (status.state === 'checking') {
    return { tone: 'info' as const, message: t('connection.checking') };
  }
  const { result } = status;
  if (result.ok) {
    return { tone: 'success' as const, message: t('connection.ok') };
  }
  const message =
    result.reason === 'http'
      ? t('connection.http', { status: result.status })
      : t(`connection.${result.reason}`);
  return { tone: 'error' as const, message };
}
