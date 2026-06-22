import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { AuthTextField } from '@/components/auth/auth-text-field';
import { PrimaryButton } from '@/components/auth/primary-button';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { Spacing } from '@/constants/theme';
import { useAuth } from '@/hooks/use-auth';

export default function ConfirmScreen() {
  const router = useRouter();
  const { phone: phoneParam } = useLocalSearchParams<{ phone?: string }>();
  const { confirmSignUp, resendCode } = useAuth();
  const [phone, setPhone] = useState(phoneParam ?? '');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);

  const onConfirm = async () => {
    if (!phone || !code) {
      Alert.alert('Missing details', 'Please enter your phone number and the code we sent you.');
      return;
    }
    setLoading(true);
    try {
      await confirmSignUp(phone, code);
      Alert.alert('Verified', 'Your account is confirmed. Please sign in.', [
        { text: 'OK', onPress: () => router.replace('/sign-in') },
      ]);
    } catch (err) {
      Alert.alert('Confirmation failed', err instanceof Error ? err.message : 'Something went wrong.');
    } finally {
      setLoading(false);
    }
  };

  const onResend = async () => {
    if (!phone) {
      Alert.alert('Phone number required', 'Enter your phone number first.');
      return;
    }
    setResending(true);
    try {
      await resendCode(phone);
      Alert.alert('Code sent', 'Check your phone for a new verification code.');
    } catch (err) {
      Alert.alert('Could not resend code', err instanceof Error ? err.message : 'Something went wrong.');
    } finally {
      setResending(false);
    }
  };

  return (
    <ThemedView style={styles.container}>
      <SafeAreaView style={styles.safeArea}>
        <ThemedText type="title" style={styles.title}>
          Verify your phone
        </ThemedText>
        <ThemedText type="small" themeColor="textSecondary" style={styles.subtitle}>
          Enter the verification code sent to your phone.
        </ThemedText>
        <ThemedView type="backgroundElement" style={styles.form}>
          <AuthTextField label="Phone number" value={phone} onChangeText={setPhone} keyboardType="phone-pad" />
          <AuthTextField label="Verification code" value={code} onChangeText={setCode} keyboardType="number-pad" />
          <PrimaryButton title="Confirm" onPress={onConfirm} loading={loading} />
        </ThemedView>
        <ThemedText type="link" style={styles.footer} onPress={onResend}>
          {resending ? 'Sending…' : 'Resend code'}
        </ThemedText>
      </SafeAreaView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  safeArea: { flex: 1, paddingHorizontal: Spacing.four, gap: Spacing.four, justifyContent: 'center' },
  title: { textAlign: 'center' },
  subtitle: { textAlign: 'center' },
  form: { gap: Spacing.three, padding: Spacing.four, borderRadius: Spacing.four },
  footer: { textAlign: 'center' },
});
