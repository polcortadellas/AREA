import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

import { readItem, storageKeys, writeItem } from '../../storage/storage';
import { parseServerUrl } from './serverUrl';

type ServerConfig = {
  /** Base URL of the AREA API, or null until the user has chosen one. */
  serverUrl: string | null;
  /** False while the saved address is being read from the device. */
  loaded: boolean;
  /** Persists a new base URL (already validated with parseServerUrl). */
  saveServerUrl: (url: string) => Promise<boolean>;
};

const ServerConfigContext = createContext<ServerConfig | null>(null);

export function ServerConfigProvider({ children }: { children: ReactNode }) {
  const [serverUrl, setServerUrl] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    let cancelled = false;
    readItem(storageKeys.serverUrl).then(saved => {
      if (cancelled) {
        return;
      }
      // Ignore anything that no longer passes validation.
      const parsed = saved === null ? null : parseServerUrl(saved);
      setServerUrl(parsed?.ok ? parsed.url : null);
      setLoaded(true);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const saveServerUrl = useCallback(async (url: string) => {
    const saved = await writeItem(storageKeys.serverUrl, url);
    if (saved) {
      setServerUrl(url);
    }
    return saved;
  }, []);

  const value = useMemo(
    () => ({ serverUrl, loaded, saveServerUrl }),
    [serverUrl, loaded, saveServerUrl],
  );

  return (
    <ServerConfigContext.Provider value={value}>
      {children}
    </ServerConfigContext.Provider>
  );
}

export function useServerConfig(): ServerConfig {
  const ctx = useContext(ServerConfigContext);
  if (!ctx) {
    throw new Error(
      'useServerConfig must be used inside <ServerConfigProvider>',
    );
  }
  return ctx;
}
