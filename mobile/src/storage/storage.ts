import AsyncStorage from '@react-native-async-storage/async-storage';

/** Every key the app persists on the device lives here. */
export const storageKeys = {
  serverUrl: 'area.serverUrl',
  language: 'area.lang',
} as const;

type StorageKey = (typeof storageKeys)[keyof typeof storageKeys];

/**
 * Small wrapper so a storage failure (full disk, corrupted DB) never crashes
 * the app: reads fall back to null and writes report false.
 */
export async function readItem(key: StorageKey): Promise<string | null> {
  try {
    return await AsyncStorage.getItem(key);
  } catch {
    return null;
  }
}

export async function writeItem(
  key: StorageKey,
  value: string,
): Promise<boolean> {
  try {
    await AsyncStorage.setItem(key, value);
    return true;
  } catch {
    return false;
  }
}
