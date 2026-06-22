import {
  AuthenticationDetails,
  CognitoRefreshToken,
  CognitoUser,
  CognitoUserAttribute,
  type CognitoUserSession,
} from 'amazon-cognito-identity-js';
import { jwtDecode } from 'jwt-decode';
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

import { toE164, userPool } from '@/lib/auth/cognito';
import { clearSession, loadStoredSession, saveSession } from '@/lib/auth/storage';

export type AuthUser = {
  sub: string;
  phone: string;
  name?: string;
};

type AuthStatus = 'loading' | 'signedIn' | 'signedOut';

type IdTokenClaims = {
  sub: string;
  phone_number?: string;
  name?: string;
};

type AuthContextValue = {
  status: AuthStatus;
  user: AuthUser | null;
  accessToken: string | null;
  signUp: (phone: string, password: string, name: string) => Promise<void>;
  confirmSignUp: (phone: string, code: string) => Promise<void>;
  resendCode: (phone: string) => Promise<void>;
  signIn: (phone: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

/** Thrown by signIn when the account exists but its phone number hasn't been confirmed yet. */
export class UnconfirmedUserError extends Error {
  phone: string;

  constructor(phone: string) {
    super('Account exists but is not confirmed yet.');
    this.phone = phone;
  }
}

function deriveSession(session: CognitoUserSession): { user: AuthUser; accessToken: string } {
  const claims = jwtDecode<IdTokenClaims>(session.getIdToken().getJwtToken());
  return {
    user: { sub: claims.sub, phone: claims.phone_number ?? '', name: claims.name },
    accessToken: session.getAccessToken().getJwtToken(),
  };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('loading');
  const [user, setUser] = useState<AuthUser | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);

  // On mount, try to resume a session from the refresh token we persisted ourselves
  // (amazon-cognito-identity-js's own storage is sync-only and unusable on React Native).
  useEffect(() => {
    let cancelled = false;

    loadStoredSession().then((stored) => {
      if (cancelled) return;
      if (!stored) {
        setStatus('signedOut');
        return;
      }

      const cognitoUser = new CognitoUser({ Username: stored.username, Pool: userPool });
      const refreshToken = new CognitoRefreshToken({ RefreshToken: stored.refreshToken });
      cognitoUser.refreshSession(refreshToken, (err, session: CognitoUserSession) => {
        if (cancelled) return;
        if (err || !session) {
          clearSession();
          setStatus('signedOut');
          return;
        }
        const { user: nextUser, accessToken: nextToken } = deriveSession(session);
        setUser(nextUser);
        setAccessToken(nextToken);
        setStatus('signedIn');
      });
    });

    return () => {
      cancelled = true;
    };
  }, []);

  const signUp = useCallback((phoneInput: string, password: string, name: string) => {
    const phone = toE164(phoneInput);
    return new Promise<void>((resolve, reject) => {
      userPool.signUp(
        phone,
        password,
        [
          new CognitoUserAttribute({ Name: 'phone_number', Value: phone }),
          new CognitoUserAttribute({ Name: 'name', Value: name }),
        ],
        [],
        (err) => (err ? reject(err) : resolve()),
      );
    });
  }, []);

  const confirmSignUp = useCallback((phoneInput: string, code: string) => {
    const cognitoUser = new CognitoUser({ Username: toE164(phoneInput), Pool: userPool });
    return new Promise<void>((resolve, reject) => {
      cognitoUser.confirmRegistration(code, true, (err) => (err ? reject(err) : resolve()));
    });
  }, []);

  const resendCode = useCallback((phoneInput: string) => {
    const cognitoUser = new CognitoUser({ Username: toE164(phoneInput), Pool: userPool });
    return new Promise<void>((resolve, reject) => {
      cognitoUser.resendConfirmationCode((err) => (err ? reject(err) : resolve()));
    });
  }, []);

  const signIn = useCallback((phoneInput: string, password: string) => {
    const phone = toE164(phoneInput);
    const cognitoUser = new CognitoUser({ Username: phone, Pool: userPool });
    const authDetails = new AuthenticationDetails({ Username: phone, Password: password });

    return new Promise<void>((resolve, reject) => {
      cognitoUser.authenticateUser(authDetails, {
        onSuccess: (session) => {
          const { user: nextUser, accessToken: nextToken } = deriveSession(session);
          saveSession({ username: phone, refreshToken: session.getRefreshToken().getToken() }).then(() => {
            setUser(nextUser);
            setAccessToken(nextToken);
            setStatus('signedIn');
            resolve();
          });
        },
        onFailure: (err) => {
          if (err?.code === 'UserNotConfirmedException') {
            reject(new UnconfirmedUserError(phone));
            return;
          }
          reject(err);
        },
      });
    });
  }, []);

  const signOut = useCallback(async () => {
    userPool.getCurrentUser()?.signOut();
    await clearSession();
    setUser(null);
    setAccessToken(null);
    setStatus('signedOut');
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ status, user, accessToken, signUp, confirmSignUp, resendCode, signIn, signOut }),
    [status, user, accessToken, signUp, confirmSignUp, resendCode, signIn, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider');
  return ctx;
}
