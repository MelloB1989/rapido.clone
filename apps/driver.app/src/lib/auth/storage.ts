import AsyncStorage from '@react-native-async-storage/async-storage';

const USERNAME_KEY = 'auth.username';
const REFRESH_TOKEN_KEY = 'auth.refreshToken';

export type StoredSession = {
  username: string;
  refreshToken: string;
};

export async function loadStoredSession(): Promise<StoredSession | null> {
  const [username, refreshToken] = await Promise.all([
    AsyncStorage.getItem(USERNAME_KEY),
    AsyncStorage.getItem(REFRESH_TOKEN_KEY),
  ]);
  if (!username || !refreshToken) return null;
  return { username, refreshToken };
}

export async function saveSession(session: StoredSession): Promise<void> {
  await Promise.all([
    AsyncStorage.setItem(USERNAME_KEY, session.username),
    AsyncStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken),
  ]);
}

export async function clearSession(): Promise<void> {
  await Promise.all([AsyncStorage.removeItem(USERNAME_KEY), AsyncStorage.removeItem(REFRESH_TOKEN_KEY)]);
}
