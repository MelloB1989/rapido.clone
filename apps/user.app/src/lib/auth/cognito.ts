import 'react-native-get-random-values';

import { CognitoUserPool } from 'amazon-cognito-identity-js';

const userPoolId = process.env.EXPO_PUBLIC_COGNITO_USER_POOL_ID;
const clientId = process.env.EXPO_PUBLIC_COGNITO_CLIENT_ID;

if (!userPoolId || !clientId) {
  throw new Error(
    'Missing EXPO_PUBLIC_COGNITO_USER_POOL_ID / EXPO_PUBLIC_COGNITO_CLIENT_ID. Copy .env.example to .env and fill in the values from the `cdk deploy` output.',
  );
}

export const userPool = new CognitoUserPool({ UserPoolId: userPoolId, ClientId: clientId });

/** Normalizes a typed phone number to E.164, defaulting to the +91 (India) country code. */
export function toE164(input: string, defaultCountryCode = '+91'): string {
  const trimmed = input.trim();
  if (trimmed.startsWith('+')) {
    return '+' + trimmed.slice(1).replace(/\D/g, '');
  }
  return defaultCountryCode + trimmed.replace(/\D/g, '');
}
